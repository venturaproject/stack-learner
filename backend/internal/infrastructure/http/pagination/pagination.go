package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Page    int64
	PerPage int64
}

func FromContext(c *gin.Context) Params {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	perPage, _ := strconv.ParseInt(c.DefaultQuery("per_page", "25"), 10, 64)
	if page < 1 {
		page = 1
	}
	if page > 2_147_483_647 {
		page = 2_147_483_647
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return Params{Page: page, PerPage: perPage}
}

func (p Params) Offset() int64 { return (p.Page - 1) * p.PerPage }
