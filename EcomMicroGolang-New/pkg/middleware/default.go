package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
)

// Key to use when setting the gin context.
const GinContextKey = "GinContextKey"

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Put the gin.Context into the request context so gqlgen can retrieve it
		ctx := context.WithValue(c.Request.Context(), GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
