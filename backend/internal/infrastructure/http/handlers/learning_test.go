package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPageQueryParsesAndBoundsHTTPParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/api/organizations?page=-3&per_page=500&search=%20academy%20", nil)
	query := pageQuery(c)
	if query.Page != 1 || query.PerPage != 25 || query.Search != " academy " {
		t.Fatalf("page query = %+v", query)
	}
}
