//go:build integration

package persistence

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	appauth "boilerplate-learner/internal/application/auth"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/http/middleware"
	"boilerplate-learner/internal/infrastructure/persistence/models"
	"boilerplate-learner/internal/shared/apperrors"

	"github.com/gin-gonic/gin"
)

func TestRefreshTokenCanOnlyBeConsumedOnceConcurrently(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewAuthRepository(db)
	jti := fmt.Sprintf("refresh-race-%d", time.Now().UnixNano())
	if err := repo.CreateRefreshToken(context.Background(), domainauth.RefreshToken{
		JTI: jti, UserID: 1, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	t.Cleanup(func() { db.Where("jti = ?", jti).Delete(&models.RefreshToken{}) })

	const contenders = 12
	start := make(chan struct{})
	type result struct {
		consumed bool
		err      error
	}
	results := make(chan result, contenders)
	var wg sync.WaitGroup
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			consumed, err := repo.ConsumeRefreshToken(context.Background(), jti, time.Now())
			results <- result{consumed: consumed, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	wins := 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("consume refresh token: %v", result.err)
		}
		if result.consumed {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("successful consumes = %d, want exactly one", wins)
	}
}

func TestDeactivatingUserRevokesRefreshSessionsAndBlocksRefresh(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewAuthRepository(db)
	user := models.User{
		Name: "Session test", Email: fmt.Sprintf("session-test-%d@example.test", time.Now().UnixNano()),
		PasswordHash: "unused", Status: "active",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		db.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})
		db.Delete(&models.User{}, user.ID)
	})
	oldJTI := fmt.Sprintf("refresh-deactivate-%d", time.Now().UnixNano())
	if err := repo.CreateRefreshToken(context.Background(), domainauth.RefreshToken{
		JTI: oldJTI, UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	userService := appauth.NewUserService(repo, repo, NewTxManager(db), NewOutboxWriter(db))
	inactive := "inactive"
	if _, err := userService.UpdateUser(context.Background(), user.ID, appauth.UpdateUserRequest{Status: &inactive}); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	storedToken, err := repo.FindRefreshTokenByJTI(context.Background(), oldJTI)
	if err != nil || storedToken == nil || !storedToken.Revoked {
		t.Fatalf("refresh token after deactivation = %+v, err=%v; want revoked", storedToken, err)
	}
	userAuthz, err := repo.UserAuthz(context.Background(), user.ID)
	if err != nil || userAuthz.Active {
		t.Fatalf("authz after deactivation = %+v, err=%v; want inactive", userAuthz, err)
	}

	gin.SetMode(gin.TestMode)
	jwt := appauth.NewJWTService(appauth.JWTConfig{Secret: "integration-test-secret", AccessTTLSecs: 900, RefreshTTLSecs: 3600})
	tokens, err := jwt.GenerateTokens(user.ID)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	router := gin.New()
	router.GET("/protected", middleware.Auth(jwt, NewRevokedTokenRepository(db), repo, []string{"admin"}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{Name: "access_token", Value: tokens.AccessToken})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("inactive user's access token status = %d, want 401", response.Code)
	}

	// Even an otherwise valid, unrevoked token cannot create a new session for
	// an inactive account (for example, one issued by a concurrent request).
	secondJTI := fmt.Sprintf("refresh-inactive-%d", time.Now().UnixNano())
	if err := repo.CreateRefreshToken(context.Background(), domainauth.RefreshToken{
		JTI: secondJTI, UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("create second refresh token: %v", err)
	}
	authService := appauth.NewAuthService(repo, repo,
		jwt,
		NewTxManager(db), NewOutboxWriter(db))
	if _, err := authService.Refresh(context.Background(), secondJTI); apperrors.KindOf(err) != apperrors.KindUnauthorized {
		t.Fatalf("refresh for inactive account error = %v, want unauthorized", err)
	}
}
