package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/weaveworks/common/server"
)

func TestGateway_RoutingAndProxying(t *testing.T) {
	// 1. Start mock upstreams
	distributor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		//nolint:gosec // G705: XSS is acceptable in test mock echo
		_, _ = fmt.Fprintf(w, "distributor:%s", r.Header.Get("X-Scope-OrgID"))
	}))
	defer distributor.Close()

	queryFrontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		//nolint:gosec // G705: XSS is acceptable in test mock echo
		_, _ = fmt.Fprintf(w, "query-frontend:%s", r.Header.Get("X-Scope-OrgID"))
	}))
	defer queryFrontend.Close()

	ruler := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		//nolint:gosec // G705: XSS is acceptable in test mock echo
		_, _ = fmt.Fprintf(w, "ruler:%s", r.Header.Get("X-Scope-OrgID"))
	}))
	defer ruler.Close()

	alertmanager := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		//nolint:gosec // G705: XSS is acceptable in test mock echo
		_, _ = fmt.Fprintf(w, "alertmanager:%s", r.Header.Get("X-Scope-OrgID"))
	}))
	defer alertmanager.Close()

	// 2. Configure Gateway
	secret := "test-secret-key-long-enough"
	cfg := Config{
		DistributorAddress:   distributor.URL,
		QueryFrontendAddress: queryFrontend.URL,
		RulerAddress:         ruler.URL,
		AlertManagerAddress:  alertmanager.URL,
		TenantIDClaim:        "tenant_id",
		JwtSecret:            secret,
	}

	serverCfg := server.Config{
		HTTPListenPort: 0,
		GRPCListenPort: 0,
	}
	svr, err := server.New(serverCfg)
	if err != nil {
		t.Fatalf("Failed to initialize weaveworks server: %v", err)
	}
	defer svr.Shutdown()

	gw, err := New(cfg, svr)
	if err != nil {
		t.Fatalf("Failed to initialize Gateway: %v", err)
	}
	gw.Start()

	// 3. Generate a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tenant_id": "mimir-tenant-123",
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	authHeader := "Bearer " + tokenString

	// 4. Test Cases
	tests := []struct {
		name           string
		method         string
		path           string
		auth           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Health check bypasses auth",
			method:         http.MethodGet,
			path:           "/health",
			auth:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   "Ok",
		},
		{
			name:           "Ingestion push (distributor)",
			method:         http.MethodPost,
			path:           "/api/v1/push",
			auth:           authHeader,
			expectedStatus: http.StatusOK,
			expectedBody:   "distributor:mimir-tenant-123",
		},
		{
			name:           "Prometheus query (query-frontend)",
			method:         http.MethodGet,
			path:           "/prometheus/api/v1/query?query=up",
			auth:           authHeader,
			expectedStatus: http.StatusOK,
			expectedBody:   "query-frontend:mimir-tenant-123",
		},
		{
			name:           "Ruler config (ruler)",
			method:         http.MethodGet,
			path:           "/prometheus/config/v1/rules",
			auth:           authHeader,
			expectedStatus: http.StatusOK,
			expectedBody:   "ruler:mimir-tenant-123",
		},
		{
			name:           "Alertmanager status (alertmanager)",
			method:         http.MethodGet,
			path:           "/alertmanager",
			auth:           authHeader,
			expectedStatus: http.StatusOK,
			expectedBody:   "alertmanager:mimir-tenant-123",
		},
		{
			name:           "Invalid token causes unauthorized",
			method:         http.MethodGet,
			path:           "/api/v1/push",
			auth:           "Bearer invalidtoken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid bearer token\n",
		},
		{
			name:           "Missing route causes not found",
			method:         http.MethodGet,
			path:           "/this-route-does-not-exist",
			auth:           "",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 - Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			rec := httptest.NewRecorder()

			svr.HTTP.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for test %s", tt.expectedStatus, rec.Code, tt.name)
			}

			body, _ := io.ReadAll(rec.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q for test %s", tt.expectedBody, string(body), tt.name)
			}
		})
	}
}
