package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Second, 5)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, time.Second, rl.interval)
	assert.Equal(t, 5, rl.burst)
}

func TestDefaultRateLimiter(t *testing.T) {
	rl := DefaultRateLimiter()
	assert.NotNil(t, rl)
	assert.Equal(t, 100, rl.rate)
	assert.Equal(t, time.Minute, rl.interval)
	assert.Equal(t, 20, rl.burst)
}

func TestRateLimiter_Allow(t *testing.T) {
	// Create limiter with burst of 3
	rl := NewRateLimiter(1, time.Second, 3)

	key := "test-client"

	// First 3 requests should be allowed (burst)
	assert.True(t, rl.Allow(key), "First request should be allowed")
	assert.True(t, rl.Allow(key), "Second request should be allowed")
	assert.True(t, rl.Allow(key), "Third request should be allowed")

	// Fourth request should be denied
	assert.False(t, rl.Allow(key), "Fourth request should be denied (burst exhausted)")
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 2)

	// Different keys should have separate buckets
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client1"))

	// client2 should still have tokens
	assert.True(t, rl.Allow("client2"))
	assert.True(t, rl.Allow("client2"))
	assert.False(t, rl.Allow("client2"))
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create limiter: 10 requests per 100ms, burst of 2
	rl := NewRateLimiter(10, 100*time.Millisecond, 2)

	key := "test-client"

	// Exhaust tokens
	assert.True(t, rl.Allow(key))
	assert.True(t, rl.Allow(key))
	assert.False(t, rl.Allow(key))

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should have tokens again
	assert.True(t, rl.Allow(key), "Should have tokens after refill")
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100, time.Second, 50)

	var wg sync.WaitGroup
	allowed := 0
	denied := 0
	mu := sync.Mutex{}

	// Make 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("concurrent-client") {
				mu.Lock()
				allowed++
				mu.Unlock()
			} else {
				mu.Lock()
				denied++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly 50 (burst capacity)
	assert.Equal(t, 50, allowed, "Should allow exactly burst capacity")
	assert.Equal(t, 50, denied, "Should deny remaining")
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 2)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := RateLimit(rl)
	wrappedHandler := middleware(handler)

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, "60", rr.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(rl)
	wrappedHandler := middleware(handler)

	// Request with X-Forwarded-For should use that IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Same forwarded IP should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	// Different forwarded IP should be allowed
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", "10.0.0.2")
	rr3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}

func TestRateLimitByKey(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Rate limit by API key header
	keyFunc := func(r *http.Request) string {
		return r.Header.Get("X-API-Key")
	}

	middleware := RateLimitByKey(rl, keyFunc)
	wrappedHandler := middleware(handler)

	// First request with API key should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "key123")
	rr1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second request with same API key should be limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "key123")
	rr2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	// Different API key should be allowed
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-API-Key", "key456")
	rr3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}
