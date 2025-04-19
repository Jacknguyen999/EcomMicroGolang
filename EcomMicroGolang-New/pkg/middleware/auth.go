package middleware

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/thomas/EcommerceAPI/pkg/auth"
)

func AuthorizeJWT(jwtService auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Always set a default empty userID
		c.Set("userID", "")

		// Check if this is a GraphQL playground request or introspection query
		if c.Request.URL.Path == "/playground" {
			c.Next()
			return
		}

		// Check if this is a login or register mutation or introspection query
		// For simplicity, we'll just check the URL path for now
		// In a real app, you'd parse the GraphQL query properly
		if c.Request.URL.Path == "/query" {
			// Read the request body
			bodyBytes, err := c.GetRawData()
			if err == nil {
				// Parse the GraphQL request
				var requestBody struct {
					Query         string                 `json:"query"`
					OperationName string                 `json:"operationName"`
					Variables     map[string]interface{} `json:"variables"`
				}

				if err := c.ShouldBindJSON(&requestBody); err == nil {
					// Check if this is a login or register operation (which don't require auth)
					if requestBody.OperationName == "Login" || requestBody.OperationName == "Register" {
						// Allow these operations without authentication
						log.Println("Allowing public operation:", requestBody.OperationName)

						// Restore the request body
						c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
						c.Next()
						return
					}

					// Allow introspection queries without authentication
					if requestBody.OperationName == "IntrospectionQuery" ||
					   (requestBody.Query != "" && (strings.Contains(requestBody.Query, "__schema") ||
					                           strings.Contains(requestBody.Query, "__type"))) {
						log.Println("Allowing introspection query without authentication")
						// Restore the request body
						c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
						c.Next()
						return
					}
				}

				// Restore the request body for later processing
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Check for Authorization header first
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		// Extract token from Authorization header if present
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
			log.Println("Found Authorization header with token")
		} else {
			// Fall back to cookie if no Authorization header
			authCookie, err := c.Cookie("token")
			if err != nil || authCookie == "" {
				// No valid token found
				log.Println("No authentication token found in header or cookie")
				// Continue with empty userID - resolvers will check and enforce auth
				c.Next()
				return
			}
			tokenString = authCookie
			log.Println("Found token in cookie")
		}

		// Validate the token
		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Println("Token validation error:", err)
			// Continue with empty userID - resolvers will check and enforce auth
			c.Next()
			return
		}

		// Extract claims from the token
		if claims, ok := token.Claims.(*auth.JWTCustomClaims); ok && token.Valid {
			log.Println("Successfully validated token")
			log.Println("User ID from token:", claims.UserID)

			// Set the userID in both gin context and request context
			c.Set("userID", claims.UserID)
			ctxWithVal := context.WithValue(c.Request.Context(), "userID", claims.UserID)
			c.Request = c.Request.WithContext(ctxWithVal)
		}

		c.Next()
	}
}
