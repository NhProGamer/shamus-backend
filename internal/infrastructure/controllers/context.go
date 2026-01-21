package controllers

import (
	"shamus-backend/internal/adapters/api/http"
	"shamus-backend/internal/adapters/api/ws"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/config"

	"github.com/coreos/go-oidc/v3/oidc"
)

// AppContext holds all application dependencies for dependency injection
type AppContext struct {
	Config           *config.Config
	GameService      ports.GameService
	WebsocketHandler *ws.WebSocketHandler
	HttpHandler      *http.GameHandler
	OIDCProvider     *oidc.Provider
}
