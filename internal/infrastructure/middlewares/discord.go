package middlewares

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DiscordAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		isAuthenticated := session.Get("authenticated")

		if isAuthenticated != true {
			// Pour une route HTTP
			if c.IsWebsocket() {
				// Si c’est une tentative de WS, on renvoie un 401
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			// Sinon, cas normal HTTP
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
