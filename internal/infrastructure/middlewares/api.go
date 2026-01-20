package middlewares

import (
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2" // Nécessaire pour wrapper le token
)

func OIDCHandler(ctx *controllers.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Récupération du token (Header ou Query)
		authContent := c.GetHeader("Authorization")
		if authContent == "" {
			authContent = c.Query("access_token")
		}

		if authContent == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token manquant"})
			return
		}

		// 2. Nettoyage (Bearer)
		rawToken := strings.TrimPrefix(authContent, "Bearer ")
		rawToken = strings.TrimSpace(rawToken)

		if rawToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token vide"})
			return
		}

		userInfo, err := ctx.OIDCProvider.UserInfo(c.Request.Context(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: rawToken,
		}))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}

		userID := userInfo.Subject

		// Injecter dans le contexte Gin
		c.Set("userID", userID)
		c.Set("userInfo", *userInfo) // Utile si vous voulez d'autres infos plus loin

		c.Next()
	}

}
