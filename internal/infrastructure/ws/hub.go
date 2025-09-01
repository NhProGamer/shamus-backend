package ws

import (
	"shamus-backend/internal/domain/entities"
	"sync"
)

type Hub struct {
	mu          sync.RWMutex
	connections map[entities.PlayerID]*ClientConn
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[entities.PlayerID]*ClientConn),
	}
}

func (h *Hub) Register(playerID entities.PlayerID, conn *ClientConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[playerID] = conn
}

func (h *Hub) Unregister(playerID entities.PlayerID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, playerID)
}

func (h *Hub) Get(playerID entities.PlayerID) (*ClientConn, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, ok := h.connections[playerID]
	return conn, ok
}
