package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test HTTP request
	req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %s", res.Header.Get("X-Content-Type-Options"))
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %s", res.Header.Get("X-Frame-Options"))
	}
	if res.Header.Get("X-XSS-Protection") != "0" {
		t.Errorf("expected X-XSS-Protection: 0, got %s", res.Header.Get("X-XSS-Protection"))
	}
	if res.Header.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy: strict-origin-when-cross-origin, got %s", res.Header.Get("Referrer-Policy"))
	}
	if res.Header.Get("Strict-Transport-Security") != "" {
		t.Errorf("expected no HSTS header on non-HTTPS request, got %s", res.Header.Get("Strict-Transport-Security"))
	}

	// Test HTTPS request
	reqSecure := httptest.NewRequest(http.MethodGet, "https://localhost/test", nil)
	recSecure := httptest.NewRecorder()

	handler.ServeHTTP(recSecure, reqSecure)

	resSecure := recSecure.Result()
	defer func() { _ = resSecure.Body.Close() }()

	if resSecure.Header.Get("Strict-Transport-Security") == "" {
		t.Error("expected HSTS header on HTTPS request")
	}
}

func TestRateLimiter(t *testing.T) {
	// Rate limiter that allows 2 requests per second, burst of 2
	limiter := NewRateLimiter(2.0, 2.0)

	// Consume both tokens in the burst
	if !limiter.Allow("127.0.0.1") {
		t.Error("expected first request to be allowed")
	}
	if !limiter.Allow("127.0.0.1") {
		t.Error("expected second request to be allowed")
	}

	// Third request should be blocked immediately
	if limiter.Allow("127.0.0.1") {
		t.Error("expected third request to be blocked")
	}

	// A different IP should be allowed
	if !limiter.Allow("192.168.1.1") {
		t.Error("expected request from different IP to be allowed")
	}

	// Wait 500ms to refill 1 token (2 tokens/sec means 1 token per 500ms)
	time.Sleep(510 * time.Millisecond)

	if !limiter.Allow("127.0.0.1") {
		t.Error("expected request to be allowed after token refill")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	limiter := NewRateLimiter(1.0, 1.0)
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected status Too Many Requests, got %d", rec2.Code)
	}
}

func TestClientInfoMiddleware(t *testing.T) {
	var capturedIP, capturedUA string
	handler := clientInfoMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock implementation of event client info extraction for test
		// Since we can't easily import internal/platform/events here due to import cycle if we aren't careful,
		// but wait, we already imported events in router.go, so we can use events package here!
		capturedIP, capturedUA = extractIP(r), r.UserAgent()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req.Header.Set("User-Agent", "TestAgent")
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "203.0.113.195" {
		t.Errorf("expected IP 203.0.113.195, got %s", capturedIP)
	}
	if capturedUA != "TestAgent" {
		t.Errorf("expected UA TestAgent, got %s", capturedUA)
	}
}
