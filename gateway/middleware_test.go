package gateway

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
	jwtReq "github.com/golang-jwt/jwt/v4/request"
)

func Test_requestContainsToken(t *testing.T) {
	tests := []struct {
		name    string
		r       *http.Request
		headers []string
		want    bool
	}{
		{
			name: "Get token using default Authorization header",
			r: &http.Request{
				Header: map[string][]string{
					"Authorization": {"Bearer: eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
				},
			},
			headers: []string{"Authorization"},
			want:    true,
		},
		{
			name: "Get token from X-Id-Token header first then Authorization (both present)",
			r: &http.Request{
				Header: map[string][]string{
					"Authorization": {"Bearer: eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
					"X-Id-Token":    {"eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
				},
			},
			headers: []string{"X-Id-Token", "Authorization"},
			want:    true,
		},
		{
			name: "Get token from X-Id-Token header first then Authorization (X-Id-Token present)",
			r: &http.Request{
				Header: map[string][]string{
					"X-Id-Token": {"eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
				},
			},
			headers: []string{"X-Id-Token", "Authorization"},
			want:    true,
		},
		{
			name: "Get token using lowercase headers",
			r: &http.Request{
				Header: map[string][]string{
					"Authorization": {"Bearer: eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
					"X-Id-Token":    {"eyJhbGciOiJIUzI1NiJ9.eyJ0ZW5hbnRfaWQiOiIxMjMiLCJ2ZXJzaW9uIjoiMSJ9.QmTSzJbvlB5_QmSmYb3nrpnUK4xuK9iWACc5xl8mmLU"},
				},
			},
			headers: []string{"x-id-token", "authorization"},
			want:    true,
		},
		{
			name:    "Missing Authorization or X-Id-Token headers",
			r:       &http.Request{},
			headers: []string{"X-Id-Token", "Authorization"},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestContainsToken(tt.r, tt.headers); got != tt.want {
				t.Errorf("requestContainsToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildHeaderExtractor(t *testing.T) {
	tests := []struct {
		name         string
		extraHeaders []string
		want         jwtReq.Extractor
	}{
		{
			name: "Default",
			want: jwtReq.MultiExtractor{
				jwtReq.AuthorizationHeaderExtractor,
			},
		},
		{
			name:         "One extra header",
			extraHeaders: []string{"X-Id-Token"},
			want: jwtReq.MultiExtractor{
				jwtReq.HeaderExtractor{"X-Id-Token"},
				jwtReq.AuthorizationHeaderExtractor,
			},
		},
		{
			name:         "Two extra headers",
			extraHeaders: []string{"X-Id-Token", "JWT-ID"},
			want: jwtReq.MultiExtractor{
				jwtReq.HeaderExtractor{"X-Id-Token"},
				jwtReq.HeaderExtractor{"JWT-ID"},
				jwtReq.AuthorizationHeaderExtractor,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildHeaderExtractor(tt.extraHeaders); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildHeaderExtractor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractTenantID(t *testing.T) {
	claim := jwt.MapClaims{
		"tenant_id":                     "1234",
		"https://example.com/tenant_id": "1234",
		"empty":                         "",
	}
	tests := []struct {
		name          string
		claim         jwt.MapClaims
		tenantIDClaim string
		want          string
		wantErr       bool
	}{
		{
			name:          "Tenant ID Claim",
			claim:         claim,
			tenantIDClaim: "tenant_id",
			want:          "1234",
			wantErr:       false,
		},
		{
			name:          "URL format Claim",
			claim:         claim,
			tenantIDClaim: "https://example.com/tenant_id",
			want:          "1234",
			wantErr:       false,
		},
		{
			name:          "Empty tenant id",
			claim:         claim,
			tenantIDClaim: "empty",
			want:          "",
			wantErr:       true,
		},
		{
			name:          "Missing Claim",
			claim:         claim,
			tenantIDClaim: "absent",
			want:          "",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTenantID(tt.claim, tt.tenantIDClaim)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTenantID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTenantID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newAuthenticationMiddleware_Vulnerability(t *testing.T) {
	// 1. Setup middleware in JWKS-only mode (JwtSecret is empty)
	cfg := Config{
		DistributorAddress:   "http://localhost:8080",
		QueryFrontendAddress: "http://localhost:8080",
		RulerAddress:         "http://localhost:8080",
		AlertManagerAddress:  "http://localhost:8080",
		TenantIDClaim:        "tenant_id",
		JwtSecret:            "",
		JwksURL:              "", // JwksURL is empty so newJWKS returns empty keys
	}

	mw := newAuthenticationMiddleware(cfg)
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:gosec // G705: XSS is acceptable in test mocks
		_, _ = w.Write([]byte(r.Header.Get("X-Scope-OrgID")))
	})
	handler := mw(innerHandler)

	// A forged token signed with HS256 and an empty key:
	// Header: {"alg":"HS256","typ":"JWT"} -> eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
	// Payload: {"tenant_id":"victim-tenant"} -> eyJ0ZW5hbnRfaWQiOiJ2aWN0aW0tdGVuYW50In0
	// HMAC-SHA256(header.payload, "") -> knCgSO_gKypWKMl-iTwBENmITbb4Aqh4tTx6ncKDzu8
	//nolint:gosec // G101: dummy token for testing
	forgedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJ2aWN0aW0tdGVuYW50In0.knCgSO_gKypWKMl-iTwBENmITbb4Aqh4tTx6ncKDzu8"

	req := httptest.NewRequest(http.MethodGet, "/prometheus/api/v1/query", nil)
	req.Header.Set("Authorization", forgedToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Errorf("Vulnerability present: request with empty-key HS256 token was accepted! Injected tenant: %q", rec.Body.String())
	} else if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized, got %d", rec.Code)
	}
}

func Test_newAuthenticationMiddleware_SymmetricValid(t *testing.T) {
	secret := "my-very-strong-secret-key-that-is-long"
	cfg := Config{
		DistributorAddress:   "http://localhost:8080",
		QueryFrontendAddress: "http://localhost:8080",
		RulerAddress:         "http://localhost:8080",
		AlertManagerAddress:  "http://localhost:8080",
		TenantIDClaim:        "tenant_id",
		JwtSecret:            secret,
	}

	mw := newAuthenticationMiddleware(cfg)
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:gosec // G705: XSS is acceptable in test mocks
		_, _ = w.Write([]byte(r.Header.Get("X-Scope-OrgID")))
	})
	handler := mw(innerHandler)

	// Create a valid HS256 token signed with the correct secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tenant_id": "valid-tenant",
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// 1. Test correct secret
	req := httptest.NewRequest(http.MethodGet, "/query", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "valid-tenant" {
		t.Errorf("Expected injected tenant 'valid-tenant', got %q", got)
	}

	// 2. Test empty secret signature should fail
	//nolint:gosec // G101: dummy token for testing
	forgedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJ2aWN0aW0tdGVuYW50In0.knCgSO_gKypWKMl-iTwBENmITbb4Aqh4tTx6ncKDzu8"
	req2 := httptest.NewRequest(http.MethodGet, "/query", nil)
	req2.Header.Set("Authorization", forgedToken)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for forged token under configured secret, got %d", rec2.Code)
	}
}
