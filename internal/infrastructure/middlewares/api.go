package middlewares

import (
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// OIDCHandler validates the access token and injects user info into context
func OIDCHandler(ctx *controllers.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header or query parameter
		authContent := c.GetHeader("Authorization")
		if authContent == "" {
			authContent = c.Query("access_token")
		}

		if authContent == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			return
		}

		// Clean up Bearer prefix
		rawToken := strings.TrimPrefix(authContent, "Bearer ")
		rawToken = strings.TrimSpace(rawToken)

		if rawToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Empty token"})
			return
		}

		// Validate token with OIDC provider
		userInfo, err := ctx.OIDCProvider.UserInfo(c.Request.Context(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: rawToken,
		}))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		userID := userInfo.Subject

		// Inject into Gin context
		c.Set("userID", userID)
		c.Set("userInfo", *userInfo)

		c.Next()
	}
}
