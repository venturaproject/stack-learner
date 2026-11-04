package auth_test

import (
	"context"
	"sync"
	"testing"
	"time"
)

// inMemoryRevokedRepo is a test double for RevokedTokenRepository.
type inMemoryRevokedRepo struct {
	mu    sync.Mutex
	store map[string]time.Time
}

func newInMemoryRevokedRepo() *inMemoryRevokedRepo {
	return &inMemoryRevokedRepo{store: make(map[string]time.Time)}
}

func (r *inMemoryRevokedRepo) Revoke(_ context.Context, jti string, exp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[jti] = exp
	return nil
}

func (r *inMemoryRevokedRepo) IsRevoked(_ context.Context, jti string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.store[jti]
	return ok, nil
}

func (r *inMemoryRevokedRepo) CleanupExpired(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	var removed int64
	for jti, exp := range r.store {
		if !exp.After(now) {
			delete(r.store, jti)
			removed++
		}
	}
	return removed, nil
}

func TestRevokedToken_UnknownJTIIsNotRevoked(t *testing.T) {
	repo := newInMemoryRevokedRepo()
	revoked, err := repo.IsRevoked(context.Background(), "nonexistent")
	if err != nil || revoked {
		t.Fatalf("expected false, nil; got %v, %v", revoked, err)
	}
}

func TestRevokedToken_RevokedJTIIsDetected(t *testing.T) {
	repo := newInMemoryRevokedRepo()
	exp := time.Now().Add(15 * time.Minute)
	_ = repo.Revoke(context.Background(), "jti-abc", exp)

	revoked, err := repo.IsRevoked(context.Background(), "jti-abc")
	if err != nil || !revoked {
		t.Fatalf("expected true, nil; got %v, %v", revoked, err)
	}
}

func TestRevokedToken_RevokeIsIdempotent(t *testing.T) {
	repo := newInMemoryRevokedRepo()
	exp := time.Now().Add(15 * time.Minute)
	_ = repo.Revoke(context.Background(), "jti-x", exp)
	_ = repo.Revoke(context.Background(), "jti-x", exp)

	revoked, _ := repo.IsRevoked(context.Background(), "jti-x")
	if !revoked {
		t.Fatal("expected revoked after duplicate revoke")
	}
}

func TestRevokedToken_CleanupRemovesExpiredKeepsActive(t *testing.T) {
	repo := newInMemoryRevokedRepo()
	_ = repo.Revoke(context.Background(), "expired", time.Unix(1, 0))
	_ = repo.Revoke(context.Background(), "active", time.Now().Add(time.Hour))

	removed, err := repo.CleanupExpired(context.Background())
	if err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}

	revoked, _ := repo.IsRevoked(context.Background(), "expired")
	if revoked {
		t.Fatal("expired token should have been cleaned up")
	}
	revoked, _ = repo.IsRevoked(context.Background(), "active")
	if !revoked {
		t.Fatal("active token should remain")
	}
}

func TestRevokedToken_DifferentJTIsAreIndependent(t *testing.T) {
	repo := newInMemoryRevokedRepo()
	_ = repo.Revoke(context.Background(), "jti-1", time.Now().Add(time.Hour))

	r1, _ := repo.IsRevoked(context.Background(), "jti-1")
	r2, _ := repo.IsRevoked(context.Background(), "jti-2")
	if !r1 || r2 {
		t.Fatalf("expected jti-1 revoked=true, jti-2 revoked=false; got %v, %v", r1, r2)
	}
}
