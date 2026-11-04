package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/infrastructure/http/middleware"
)

func idemApp(hits *int32) *gin.Engine {
	r := gin.New()
	r.Use(middleware.IdempotencyKey(nil)) // nil Redis → in-process store
	h := func(c *gin.Context) {
		atomic.AddInt32(hits, 1)
		c.JSON(http.StatusCreated, gin.H{"id": atomic.LoadInt32(hits)})
	}
	r.POST("/things", h)
	r.GET("/things", h)
	return r
}

func post(r *gin.Engine, key string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/things", nil)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIdempotency_ReplaysSameKey(t *testing.T) {
	var hits int32
	r := idemApp(&hits)

	first := post(r, "abc-123")
	second := post(r, "abc-123")

	if hits != 1 {
		t.Fatalf("handler ran %d times, want 1", hits)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("replay body mismatch: %q vs %q", first.Body.String(), second.Body.String())
	}
	if second.Code != http.StatusCreated {
		t.Fatalf("replay status = %d, want 201", second.Code)
	}
	if second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("replay must set Idempotency-Replayed header")
	}
}

func TestIdempotency_DifferentKeysBothRun(t *testing.T) {
	var hits int32
	r := idemApp(&hits)
	post(r, "key-1")
	post(r, "key-2")
	if hits != 2 {
		t.Fatalf("handler ran %d times, want 2", hits)
	}
}

func TestIdempotency_NoKeyAlwaysRuns(t *testing.T) {
	var hits int32
	r := idemApp(&hits)
	post(r, "")
	post(r, "")
	if hits != 2 {
		t.Fatalf("handler ran %d times, want 2", hits)
	}
}

func TestIdempotency_IgnoresNonPost(t *testing.T) {
	var hits int32
	r := idemApp(&hits)
	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/things", nil)
		req.Header.Set("Idempotency-Key", "same")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
	if hits != 2 {
		t.Fatalf("GET ran %d times, want 2 (idempotency must not touch GET)", hits)
	}
}

func TestIdempotency_RejectsOversizedKey(t *testing.T) {
	var hits int32
	r := idemApp(&hits)
	w := post(r, strings.Repeat("x", 300))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if hits != 0 {
		t.Fatal("handler must not run for an invalid key")
	}
}
