package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("192.168.1.1") {
		t.Error("4th request should be denied")
	}

	// Different IP should be allowed
	if !rl.Allow("192.168.1.2") {
		t.Error("different IP should be allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(1, 100*time.Millisecond)

	// First request should be allowed
	if !rl.Allow("192.168.1.1") {
		t.Error("first request should be allowed")
	}

	// Second request should be denied
	if rl.Allow("192.168.1.1") {
		t.Error("second request should be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed again
	if !rl.Allow("192.168.1.1") {
		t.Error("request after window expiry should be allowed")
	}
}

func TestCORSMiddleware(t *testing.T) {
	cfg := DefaultCORSConfig()
	middleware := CORS(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("expected CORS origin '*', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
		}

		if rec.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("expected Access-Control-Allow-Methods header")
		}
	})

	t.Run("regular request with origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("expected CORS origin '*', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("request without origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	middleware := CORS(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("expected origin 'http://localhost:3000', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
		}

		if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("expected credentials header")
		}
	})

	t.Run("disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://evil.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("expected no CORS header for disallowed origin, got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	middleware := RateLimitMiddleware(rl)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should pass
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d should be allowed, got status %d", i+1, rec.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single",
			remoteAddr: "192.168.1.1:12345",
			xff:        "10.0.0.1",
			expected:   "10.0.0.1",
		},
		{
			name:       "X-Forwarded-For multiple",
			remoteAddr: "192.168.1.1:12345",
			xff:        "10.0.0.1, 10.0.0.2",
			expected:   "10.0.0.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "192.168.1.1:12345",
			xri:        "10.0.0.1",
			expected:   "10.0.0.1",
		},
		{
			name:       "RemoteAddr with port",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("join", func(t *testing.T) {
		result := join([]string{"a", "b", "c"}, ", ")
		if result != "a, b, c" {
			t.Errorf("expected 'a, b, c', got '%s'", result)
		}

		result = join([]string{}, ", ")
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}

		result = join([]string{"single"}, ", ")
		if result != "single" {
			t.Errorf("expected 'single', got '%s'", result)
		}
	})

	t.Run("itoa", func(t *testing.T) {
		tests := []struct {
			input    int
			expected string
		}{
			{0, "0"},
			{1, "1"},
			{10, "10"},
			{123, "123"},
			{86400, "86400"},
		}

		for _, tt := range tests {
			result := itoa(tt.input)
			if result != tt.expected {
				t.Errorf("itoa(%d) = %s, expected %s", tt.input, result, tt.expected)
			}
		}
	})
}
