package controllers

import (
	"shamus-backend/internal/adapters/api/http"
	"shamus-backend/internal/adapters/api/ws"
	"shamus-backend/internal/adapters/app_adapters"
	"shamus-backend/internal/infrastructure/config"

	"github.com/coreos/go-oidc/v3/oidc"
)

type AppContext struct {
	Config           *config.Config
	GameService      *app_adapters.GameService
	WebsocketHandler *ws.WebSocketHandler
	HttpHandler      *http.GameHandler
	OIDCProvider     *oidc.Provider
}
