package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserId(ctx context.Context, abort bool) string {
	accountId, ok := ctx.Value("userID").(string)
	log.Println("userID", accountId)
	if !ok || accountId == "" {
		log.Println("Authentication failed: No valid user ID in context")
		if abort {
			ginContext, _ := ctx.Value("GinContextKey").(*gin.Context)
			if ginContext != nil {
				log.Println("Aborting request due to authentication failure")
				ginContext.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized: You must be logged in to perform this action",
				})
			} else {
				log.Println("Warning: Could not abort request, gin context not found")
			}
		}
		return ""
	}
	return accountId
}

func GetUserIdInt(ctx context.Context, abort bool) (int, error) {
	idString := GetUserId(ctx, abort)
	log.Println("userID", idString)
	if idString == "" {
		return 0, errors.New("unauthorized: you must be logged in to perform this action")
	}

	idInt, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Println("Error parsing user ID:", err)
		return 0, errors.New("invalid user ID format")
	}

	return int(idInt), nil
}
