package controllers

import (
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

// WebSocketHandler defines the interface for WebSocket handlers
type WebSocketHandler interface {
	HandleWS(c *gin.Context)
}

// AppContext holds all application dependencies for dependency injection
type AppContext struct {
	Config           *config.Config
	GameService      ports.GameService
	WebsocketHandler WebSocketHandler
	OIDCProvider     *oidc.Provider
}
