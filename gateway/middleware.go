package gateway

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/cortexproject/cortex/pkg/util/log"
	klog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	jwt "github.com/golang-jwt/jwt/v4"
	jwtReq "github.com/golang-jwt/jwt/v4/request"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/middleware"
)

var (
	jwtSecret       string
	extraHeadersArg string
	tenantName      string
	tenantIDClaim   string

	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "failed_authentications_total",
		Help:      "The total number of failed authentications.",
	}, []string{"reason"})
	authSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "succeeded_authentications_total",
		Help:      "The total number of succeeded authentications.",
	}, []string{"tenant"})
)

func init() {
	flag.StringVar(&tenantName, "gateway.auth.tenant-name", "", "Tenant name to use when jwt auth disabled")
	flag.StringVar(&jwtSecret, "gateway.auth.jwt-secret", "", "Secret to sign JSON Web Tokens")
	flag.StringVar(&extraHeadersArg, "gateway.auth.jwt-extra-headers", "", "A comma separated list of additional headers to scan for JSON Web Tokens presence")
	flag.StringVar(&tenantIDClaim, "gateway.auth.tenant-id-claim", "tenant_id", "The name of the Tenant ID Claim. Defaults to tenant_id")
}

// AuthenticateTenant validates the Bearer Token and attaches the TenantID to the request
var AuthenticateTenant = middleware.Func(func(next http.Handler) http.Handler {

	var extraHeaders []string
	if extraHeadersArg != "" {
		extraHeaders = strings.Split(extraHeadersArg, ",")
	}
	headers := append(extraHeaders, "Authorization")
	authorizationHeaderExtractor := buildHeaderExtractor(extraHeaders)
	jwks := newJWKS()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if tenantName != "" {
			r.Header.Set("X-Scope-OrgID", tenantName)
			next.ServeHTTP(w, r)
			return
		}
		logger := klog.With(log.WithContext(r.Context(), log.Logger), "ip_address", r.RemoteAddr)
		level.Debug(logger).Log("msg", "authenticating request", "route", r.RequestURI)

		if !requestContainsToken(r, headers) {
			level.Info(logger).Log("msg", "no bearer token provided")
			http.Error(w, "No bearer token provided", http.StatusUnauthorized)
			authFailures.WithLabelValues("no_token").Inc()
			return
		}

		// Try to parse and validate JWT
		te := jwt.MapClaims{}
		_, err := jwtReq.ParseFromRequest(
			r,
			authorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				keyAlg := token.Method.Alg()
				switch keyAlg {
				case "RS256", "RS384", "RS512", "EdDSA", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512":
					return jwks.Keyfunc(token)
				case "HS256", "HS384", "HS512":
					return []byte(jwtSecret), nil
				}
				return nil, fmt.Errorf("Unexpected signing method: %v", keyAlg)
			},
			jwtReq.WithClaims(te))

		// If Tenant's Valid method returns false an error will be set as well, hence there is no need
		// to additionally check the parsed token for "Valid"
		if err != nil {
			level.Info(logger).Log("msg", "invalid bearer token", "err", err.Error())
			http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
			authFailures.WithLabelValues("token_not_valid").Inc()
			return
		}

		tenantID, err := extractTenantID(te, tenantIDClaim)
		if err != nil {
			level.Info(logger).Log("msg", "invalid tenant id", "err", err.Error())
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

func extractTenantID(claim jwt.MapClaims, tenantIDClaim string) (string, error) {
	tenantID, tenantIDClaimFound := claim[tenantIDClaim]
	if !tenantIDClaimFound {
		return "", fmt.Errorf("Claim %v not found", tenantIDClaim)
	}
	tenantIDStr := tenantID.(string)
	if tenantIDStr == "" {
		return "", fmt.Errorf("Empty Tenant ID")
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
