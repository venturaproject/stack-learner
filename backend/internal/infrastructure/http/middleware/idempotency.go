package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/infrastructure/cache"
)

const (
	idempotencyHeader       = "Idempotency-Key"
	idempotencyLockTTL      = 2 * time.Minute
	idempotencyResultTTL    = 24 * time.Hour
	idempotencyMaxKeyLen    = 255
	idempotencyLockSentinel = "\x00lock"
)

// idempotencyRecord is the stored response of a completed request.
type idempotencyRecord struct {
	Status  int               `json:"status"`
	Body    []byte            `json:"body"`
	Headers map[string]string `json:"headers"`
}

// replayHeaders is the small allowlist of response headers worth persisting so a
// replayed response is byte-equivalent to the original.
var replayHeaders = []string{"Content-Type", "Location"}

// IdempotencyKey makes unsafe requests safe to retry. It is opt-in: a client
// sends an `Idempotency-Key` header (a unique token, typically a UUID) and the
// first request with that key runs normally while its response is cached; any
// retry with the same key replays the stored response instead of executing the
// handler again. Only POST is intercepted — PUT/PATCH/DELETE are already
// idempotent by contract, and GET/HEAD have nothing to dedupe.
//
// With Redis the cache is shared across processes; without it (or if Redis
// errors) the middleware degrades to a best-effort in-process cache.
func IdempotencyKey(redis *cache.Redis) gin.HandlerFunc {
	var store idempotencyStore = newMemIdempotencyStore()
	if redis != nil {
		store = redisIdempotencyStore{r: redis}
	}

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		key := c.GetHeader(idempotencyHeader)
		if key == "" {
			c.Next()
			return
		}
		if len(key) > idempotencyMaxKeyLen {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key demasiado largo"})
			return
		}

		ctx := c.Request.Context()
		storeKey := "idem:" + clientIP(c) + ":" + c.Request.Method + ":" + c.FullPath() + ":" + key

		rec, claimed, err := store.begin(ctx, storeKey, idempotencyLockTTL)
		switch {
		case err != nil:
			// Cache unavailable — proceed without dedupe rather than fail the request.
			c.Next()
			return
		case rec != nil:
			replay(c, *rec)
			return
		case !claimed:
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "solicitud idéntica en curso"})
			return
		}

		buf := &bytes.Buffer{}
		c.Writer = &bodyCapture{ResponseWriter: c.Writer, buf: buf}

		c.Next()

		status := c.Writer.Status()
		if status >= http.StatusInternalServerError || len(c.Errors) > 0 {
			// Transient failure: drop the claim so the client can retry cleanly.
			_ = store.release(ctx, storeKey)
			return
		}

		headers := make(map[string]string)
		for _, h := range replayHeaders {
			if v := c.Writer.Header().Get(h); v != "" {
				headers[h] = v
			}
		}
		_ = store.complete(ctx, storeKey, idempotencyRecord{
			Status:  status,
			Body:    buf.Bytes(),
			Headers: headers,
		}, idempotencyResultTTL)
	}
}

func replay(c *gin.Context, rec idempotencyRecord) {
	for k, v := range rec.Headers {
		c.Header(k, v)
	}
	c.Header("Idempotency-Replayed", "true")
	c.Data(rec.Status, rec.Headers["Content-Type"], rec.Body)
	c.Abort()
}

type bodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyCapture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyCapture) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// idempotencyStore claims a key, replays a finished response, or reports that a
// concurrent request already holds the key.
//
//   - begin: (rec!=nil) a finished response is cached — replay it.
//     (rec==nil, claimed) the caller owns the slot; it must complete or release.
//     (rec==nil, !claimed) another request is in flight — return 409.
type idempotencyStore interface {
	begin(ctx context.Context, key string, lockTTL time.Duration) (rec *idempotencyRecord, claimed bool, err error)
	complete(ctx context.Context, key string, rec idempotencyRecord, ttl time.Duration) error
	release(ctx context.Context, key string) error
}

// ── Redis-backed ─────────────────────────────────────────────────────────────

type redisIdempotencyStore struct{ r *cache.Redis }

func (s redisIdempotencyStore) begin(ctx context.Context, key string, lockTTL time.Duration) (*idempotencyRecord, bool, error) {
	ok, err := s.r.SetNX(ctx, key, idempotencyLockSentinel, lockTTL)
	if err != nil {
		return nil, false, err
	}
	if ok {
		return nil, true, nil
	}
	raw, err := s.r.Get(ctx, key)
	if err != nil || raw == idempotencyLockSentinel {
		return nil, false, nil // in flight (or vanished mid-race)
	}
	var rec idempotencyRecord
	if json.Unmarshal([]byte(raw), &rec) != nil {
		return nil, false, nil
	}
	return &rec, false, nil
}

func (s redisIdempotencyStore) complete(ctx context.Context, key string, rec idempotencyRecord, ttl time.Duration) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.r.Set(ctx, key, payload, ttl)
}

func (s redisIdempotencyStore) release(ctx context.Context, key string) error {
	return s.r.Del(ctx, key)
}

// ── In-process fallback ──────────────────────────────────────────────────────

type memIdempotencyEntry struct {
	rec       *idempotencyRecord
	expiresAt time.Time
}

type memIdempotencyStore struct {
	mu sync.Mutex
	m  map[string]memIdempotencyEntry
}

func newMemIdempotencyStore() *memIdempotencyStore {
	s := &memIdempotencyStore{m: make(map[string]memIdempotencyEntry)}
	go s.sweep()
	return s
}

func (s *memIdempotencyStore) begin(_ context.Context, key string, lockTTL time.Duration) (*idempotencyRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.m[key]; ok && time.Now().Before(e.expiresAt) {
		if e.rec != nil {
			return e.rec, false, nil
		}
		return nil, false, nil
	}
	s.m[key] = memIdempotencyEntry{expiresAt: time.Now().Add(lockTTL)}
	return nil, true, nil
}

func (s *memIdempotencyStore) complete(_ context.Context, key string, rec idempotencyRecord, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := rec
	s.m[key] = memIdempotencyEntry{rec: &stored, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (s *memIdempotencyStore) release(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	return nil
}

func (s *memIdempotencyStore) sweep() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, e := range s.m {
			if now.After(e.expiresAt) {
				delete(s.m, k)
			}
		}
		s.mu.Unlock()
	}
}
