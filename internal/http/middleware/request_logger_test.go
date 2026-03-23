package middleware

import "testing"

func TestShouldSkipRequestLog_AnonymousAuthBootstrapNoise(t *testing.T) {
	if !shouldSkipRequestLog("GET", "/v2/auth/me", "/v2/auth/me", 401) {
		t.Fatalf("expected /v2/auth/me 401 to be skipped")
	}
	if !shouldSkipRequestLog("POST", "/v2/auth/refresh", "/v2/auth/refresh", 204) {
		t.Fatalf("expected /v2/auth/refresh 204 to be skipped")
	}
	if shouldSkipRequestLog("POST", "/v2/auth/login", "/v2/auth/login", 401) {
		t.Fatalf("did not expect /v2/auth/login 401 to be skipped")
	}
}
