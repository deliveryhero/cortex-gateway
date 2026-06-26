package gateway

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/cortexproject/cortex/pkg/util/log"
	klog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	jwt "github.com/golang-jwt/jwt/v4"
	jwtReq "github.com/golang-jwt/jwt/v4/request"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/middleware"
)

const metricsNamespace = "cortex_gateway"

var (
	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "failed_authentications_total",
		Help:      "The total number of failed authentications.",
	}, []string{"reason"})
	authSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "succeeded_authentications_total",
		Help:      "The total number of succeeded authentications.",
	}, []string{"tenant"})
)

// AuthenticateTenant validates the Bearer Token and attaches the TenantID to the request
func newAuthenticationMiddleware(cfg Config) middleware.Func {
	if cfg.TenantName != "" {
		return newStaticTenantNameMiddleware(cfg.TenantName)
	}

	var extraHeaders []string
	if cfg.ExtraHeaders != "" {
		extraHeaders = strings.Split(cfg.ExtraHeaders, ",")
	}
	headers := make([]string, 0, len(extraHeaders)+1)
	headers = append(headers, extraHeaders...)
	headers = append(headers, "Authorization")
	authorizationHeaderExtractor := buildHeaderExtractor(extraHeaders)
	jwks := newJWKS(cfg)

	var validMethods []string
	if cfg.JwksURL != "" {
		validMethods = append(validMethods, "RS256", "RS384", "RS512", "EdDSA", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512")
	}
	if cfg.JwtSecret != "" {
		validMethods = append(validMethods, "HS256", "HS384", "HS512")
	}

	return middleware.Func(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := klog.With(log.WithContext(r.Context(), log.Logger), "ip_address", r.RemoteAddr)
			_ = level.Debug(logger).Log("msg", "authenticating request", "route", r.RequestURI)

			if !requestContainsToken(r, headers) {
				_ = level.Info(logger).Log("msg", "no bearer token provided")
				http.Error(w, "No bearer token provided", http.StatusUnauthorized)
				authFailures.WithLabelValues("no_token").Inc()
				return
			}

			// Try to parse and validate JWT
			te := jwt.MapClaims{}
			parser := &jwt.Parser{
				ValidMethods: validMethods,
			}
			_, err := jwtReq.ParseFromRequest(
				r,
				authorizationHeaderExtractor,
				newKeyfunc(cfg.JwtSecret, jwks),
				jwtReq.WithClaims(te),
				jwtReq.WithParser(parser))

			// If Tenant's Valid method returns false an error will be set as well, hence there is no need
			// to additionally check the parsed token for "Valid"
			if err != nil {
				_ = level.Info(logger).Log("msg", "invalid bearer token", "err", err.Error())
				http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
				authFailures.WithLabelValues("token_not_valid").Inc()
				return
			}

			tenantID, err := extractTenantID(te, cfg.TenantIDClaim)
			if err != nil {
				_ = level.Info(logger).Log("msg", "invalid tenant id", "err", err.Error())
				http.Error(w, "Invalid Tenant ID", http.StatusUnauthorized)
				authFailures.WithLabelValues("tenant_id_not_valid").Inc()
				return
			}

			// Token is valid
			authSuccess.WithLabelValues(tenantID).Inc()
			r.Header.Set("X-Scope-OrgID", tenantID)
			next.ServeHTTP(w, r)
		})
	})
}

func newStaticTenantNameMiddleware(tenantName string) middleware.Func {
	return middleware.Func(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Scope-OrgID", tenantName)
			next.ServeHTTP(w, r)
		})
	})
}

func newKeyfunc(jwtSecret string, jwks *keyfunc.JWKS) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		keyAlg := token.Method.Alg()
		switch keyAlg {
		case "RS256", "RS384", "RS512", "EdDSA", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512":
			return jwks.Keyfunc(token)
		case "HS256", "HS384", "HS512":
			if jwtSecret == "" {
				return nil, fmt.Errorf("symmetric signing method %v is not configured", keyAlg)
			}
			return []byte(jwtSecret), nil
		}
		return nil, fmt.Errorf("unexpected signing method: %v", keyAlg)
	}
}

func extractTenantID(claim jwt.MapClaims, tenantIDClaim string) (string, error) {
	tenantID, tenantIDClaimFound := claim[tenantIDClaim]
	if !tenantIDClaimFound {
		return "", fmt.Errorf("claim %v not found", tenantIDClaim)
	}
	tenantIDStr, ok := tenantID.(string)
	if !ok {
		return "", fmt.Errorf("tenant id claim is not a string")
	}
	if tenantIDStr == "" {
		return "", fmt.Errorf("empty tenant id")
	}
	return tenantIDStr, nil
}

func buildHeaderExtractor(extraHeaders []string) jwtReq.Extractor {
	authorizationHeaderExtractor := make(jwtReq.MultiExtractor, len(extraHeaders)+1)
	for i, header := range extraHeaders {
		authorizationHeaderExtractor[i] = jwtReq.HeaderExtractor{header}
	}
	authorizationHeaderExtractor[len(extraHeaders)] = jwtReq.AuthorizationHeaderExtractor
	return authorizationHeaderExtractor
}

func requestContainsToken(r *http.Request, headers []string) bool {
	for _, header := range headers {
		idToken := r.Header.Get(header)
		if idToken != "" {
			return true
		}
	}
	return false
}
