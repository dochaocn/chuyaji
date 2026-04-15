package middleware

import (
	"net/http"
	"strings"

	"github.com/dochaocn/chuyaji/services/api/internal/auth"
	"github.com/gin-gonic/gin"
)

const ctxUserIDKey = "uid"

// UserID returns the authenticated user id set by JWT middleware.
func UserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ctxUserIDKey)
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case uint64:
		return x, true
	case float64:
		return uint64(x), true
	default:
		return 0, false
	}
}

func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		cl, err := auth.ParseJWT(secret, raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(ctxUserIDKey, cl.UserID)
		c.Next()
	}
}
