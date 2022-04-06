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
	flag.StringVar(&jwtSecret, "gateway.auth.jwt-secret", "", "Secret to sign JSON Web Tokens")
	flag.StringVar(&extraHeadersArg, "gateway.auth.jwt-extra-headers", "", "A comma separated list of additional headers to scan for JSON Web Tokens presence")
}

// AuthenticateTenant validates the Bearer Token and attaches the TenantID to the request
var AuthenticateTenant = middleware.Func(func(next http.Handler) http.Handler {

	var extraHeaders []string
	if extraHeadersArg != "" {
		extraHeaders = strings.Split(extraHeadersArg, ",")
	}
	headers := append(extraHeaders, "Authorization")
	authorizationHeaderExtractor := buildHeaderExtractor(extraHeaders)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := klog.With(log.WithContext(r.Context(), log.Logger), "ip_address", r.RemoteAddr)
		level.Debug(logger).Log("msg", "authenticating request", "route", r.RequestURI)

		if !requestContainsToken(r, headers) {
			level.Info(logger).Log("msg", "no bearer token provided")
			http.Error(w, "No bearer token provided", http.StatusUnauthorized)
			authFailures.WithLabelValues("no_token").Inc()
			return
		}

		// Try to parse and validate JWT
		te := &tenant{}
		_, err := jwtReq.ParseFromRequest(
			r,
			authorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				// Only HMAC algorithms accepted - algorithm validation is super important!
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					level.Info(logger).Log("msg", "unexpected signing method", "used_method", token.Header["alg"])
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(jwtSecret), nil
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

		// Token is valid
		authSuccess.WithLabelValues(te.TenantID).Inc()
		r.Header.Set("X-Scope-OrgID", te.TenantID)
		next.ServeHTTP(w, r)
	})
})

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
