package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var isAdmin bool
		a, ok := c.Get("isAdmin")
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		isAdmin, _ = a.(bool)
		if !isAdmin {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
