package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func pathInt32(c *gin.Context, param string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(param), 10, 32)
	return int32(v), err
}

func queryStr(c *gin.Context, key string) *string {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	return &v
}
