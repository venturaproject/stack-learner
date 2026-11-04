//go:build integration

package persistence

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"boilerplate-learner/internal/application/learning"
	learningread "boilerplate-learner/internal/application/learning/read"
	"boilerplate-learner/internal/domain/organizations"
	"boilerplate-learner/internal/infrastructure/http/handlers"
	"boilerplate-learner/internal/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

func TestListOrganizationsHTTPContractIncludesPagination(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ctx := t.Context()
	userID := int32(time.Now().UnixNano() % 1_000_000_000)
	repo := NewLearningRepository(db)
	service := learning.NewService(repo, NewTxManager(db))
	for _, name := range []string{"Target contract org", "Other contract org"} {
		if _, err := service.CreateOrganization(ctx, userID, name, organizations.NormalizeSlug(fmt.Sprintf("%s-%d", name, time.Now().UnixNano()))); err != nil {
			t.Fatalf("create organization: %v", err)
		}
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM organization_memberships WHERE user_id = ?", userID)
		db.Exec("DELETE FROM organizations WHERE created_by = ?", userID)
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := handlers.NewLearningHandler(service)
	router.GET("/organizations", func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		handler.ListOrganizations(c)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/organizations?page=1&per_page=1&search=Target", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Organizations []learningread.Organization `json:"organizations"`
		Pagination    struct {
			Page       int   `json:"page"`
			PerPage    int   `json:"per_page"`
			Total      int64 `json:"total"`
			TotalPages int   `json:"total_pages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Organizations) != 1 || body.Organizations[0].Name != "Target contract org" || body.Organizations[0].Role != organizations.RoleAdmin {
		t.Fatalf("organizations response = %+v", body.Organizations)
	}
	if body.Pagination.Page != 1 || body.Pagination.PerPage != 1 || body.Pagination.Total != 1 || body.Pagination.TotalPages != 1 {
		t.Fatalf("pagination response = %+v", body.Pagination)
	}
}
