package middlewares

import (
	"log"
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type UserClaims struct {
	Sub string `json:"sub"`
}

func OIDCHandler(ctx *controllers.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		oidcConfig := &oidc.Config{
			ClientID: ctx.Config.OIDC.ClientID,
		}
		verifier := ctx.OIDCProvider.Verifier(oidcConfig)
		// 1. Récupération du contenu brut (Header ou Fallback Query Param)
		authContent := c.GetHeader("Authorization")
		if authContent == "" {
			authContent = c.Query("access_token")
		}

		// 2. Vérification si vide après le fallback
		if authContent == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header ou paramètre access_token manquant"})
			return
		}

		// 3. Nettoyage pour obtenir uniquement le token (sans "Bearer ")
		// La variable finale s'appelle 'rawToken'
		rawToken := strings.TrimPrefix(authContent, "Bearer ")
		rawToken = strings.TrimSpace(rawToken)

		if rawToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Format de token invalide"})
			return
		}

		// À partir d'ici, vous avez 'rawToken' (string)

		// Vérification du token via OIDC
		// Cette étape vérifie la signature, l'expiration et l'émetteur
		idToken, err := verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			log.Println(rawToken)
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}

		// Extraction des claims (infos utilisateur)
		var claims UserClaims
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Impossible de lire les claims"})
			return
		}

		// Injecter les infos dans le contexte Gin pour les handlers suivants
		c.Set("userID", claims.Sub)
		c.Next()
	}
}
