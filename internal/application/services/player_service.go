package services

import (
	"context"
	"encoding/json"
	"math/rand"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
	"sync"
	"time"
)

// ConnectionChecker is an interface to check if a player has an active WebSocket session
// This avoids circular dependency with WebSocketHandler
type ConnectionChecker interface {
	IsPlayerConnected(playerID entities.PlayerID) bool
	BroadcastToGame(gameID entities.GameID, payload []byte) error
}

// HostChangedNotifier is an interface to notify host changes
// This avoids circular dependency with NotificationService
type HostChangedNotifier interface {
	NotifyHostChanged(gameID entities.GameID, newHostID entities.PlayerID, newHostUsername string) error
}

type PlayerService struct {
	playerRepo   ports.PlayerRepository
	gameRepo     ports.GameRepository
	connChecker  ConnectionChecker
	notifier     HostChangedNotifier
	reconnTimers map[entities.PlayerID]*time.Timer
	timerLock    sync.Mutex
}

func NewPlayerService(
	playerRepo ports.PlayerRepository,
	gameRepo ports.GameRepository,
	connChecker ConnectionChecker,
) *PlayerService {
	return &PlayerService{
		playerRepo:   playerRepo,
		gameRepo:     gameRepo,
		connChecker:  connChecker,
		reconnTimers: make(map[entities.PlayerID]*time.Timer),
	}
}

// SetNotifier sets the host changed notifier (to break circular dependency)
func (s *PlayerService) SetNotifier(notifier HostChangedNotifier) {
	s.notifier = notifier
}

// HandleConnect is called when a player connects via WebSocket
// Returns the Player and a boolean indicating if this is a reconnection
func (s *PlayerService) HandleConnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID, username string) (*entities.Player, bool, error) {
	// 1. Check if player is already connected (active WebSocket session)
	if s.connChecker != nil && s.connChecker.IsPlayerConnected(playerID) {
		return nil, false, apperrors.ErrPlayerAlreadyConnected
	}

	// 2. Get the game
	game, err := s.gameRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, false, apperrors.ErrGameNotFound
	}

	// 3. Check if game has ended - no one can join
	if game.Status == entities.GameStatusEnded {
		return nil, false, apperrors.ErrGameEnded
	}

	// 4. Check if player already exists
	existingPlayer, err := s.playerRepo.GetPlayer(ctx, playerID)

	if err != nil {
		// Player not found - this is a new player
		if game.Status != entities.GameStatusWaiting {
			return nil, false, apperrors.ErrGameNotWaiting
		}

		// Create new player
		player := entities.NewPlayer(playerID, username, &gameID)
		if err := s.playerRepo.SavePlayer(ctx, player); err != nil {
			return nil, false, err
		}

		// Add to game's player set
		if err := s.playerRepo.AddPlayerToGame(ctx, gameID, playerID); err != nil {
			return nil, false, err
		}

		// Add to game.Players slice if not already present
		found := false
		for _, p := range game.Players {
			if p == playerID {
				found = true
				break
			}
		}
		if !found {
			game.Players = append(game.Players, playerID)
			if err := s.gameRepo.SaveGame(ctx, game); err != nil {
				return nil, false, err
			}
		}

		return player, false, nil // false = not a reconnection
	}

	// Player exists - check their state
	if existingPlayer.ConnectionState == entities.ConnectionStateInactive {
		return nil, false, apperrors.ErrPlayerInactive
	}

	// Check if player belongs to this game
	if existingPlayer.GameID == nil || *existingPlayer.GameID != gameID {
		// Player exists but for a different game
		if game.Status != entities.GameStatusWaiting {
			return nil, false, apperrors.ErrGameNotWaiting
		}

		// Update player's game
		existingPlayer.GameID = &gameID
		existingPlayer.Connect()
		if err := s.playerRepo.SavePlayer(ctx, existingPlayer); err != nil {
			return nil, false, err
		}

		// Add to this game's player set
		if err := s.playerRepo.AddPlayerToGame(ctx, gameID, playerID); err != nil {
			return nil, false, err
		}

		return existingPlayer, false, nil
	}

	// Player is reconnecting to the same game
	// Only allow reconnection if game is active and player was disconnected
	if game.Status == entities.GameStatusActive {
		if existingPlayer.ConnectionState != entities.ConnectionStateDisconnected {
			// Player is marked as connected but trying to connect again
			// This shouldn't happen if IsPlayerConnected works correctly
			return nil, false, apperrors.ErrPlayerAlreadyConnected
		}
	}

	s.cancelReconnTimer(playerID)
	existingPlayer.Connect()
	if err := s.playerRepo.SavePlayer(ctx, existingPlayer); err != nil {
		return nil, false, err
	}

	return existingPlayer, true, nil // true = reconnection
}

// HandleDisconnect is called when a player disconnects from WebSocket
func (s *PlayerService) HandleDisconnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error {
	game, err := s.gameRepo.GetGame(ctx, gameID)
	if err != nil {
		return err
	}

	player, err := s.playerRepo.GetPlayer(ctx, playerID)
	if err != nil {
		return err
	}

	switch game.Status {
	case entities.GameStatusWaiting:
		// Remove player completely from the game
		if err := s.playerRepo.RemovePlayerFromGame(ctx, gameID, playerID); err != nil {
			return err
		}
		if err := s.playerRepo.DeletePlayer(ctx, playerID); err != nil {
			return err
		}

		// Remove from game.Players slice
		game.Players = removePlayerFromSlice(game.Players, playerID)

		// Host migration: if the leaving player was the host, assign a new random host
		if game.HostID == playerID && len(game.Players) > 0 {
			newHostID := selectRandomHost(game.Players)
			game.HostID = newHostID

			// Get new host's username for notification
			newHost, err := s.playerRepo.GetPlayer(ctx, newHostID)
			var newHostUsername string
			if err == nil && newHost != nil {
				newHostUsername = newHost.Username
			}

			logger.Get().Info().
				Str("gameID", string(gameID)).
				Str("oldHost", string(playerID)).
				Str("newHost", string(newHostID)).
				Msg("Host migrated to new player")

			// Notify all players of host change (after saving game)
			defer func() {
				if s.notifier != nil {
					s.notifier.NotifyHostChanged(gameID, newHostID, newHostUsername)
				}
			}()
		}

		if err := s.gameRepo.SaveGame(ctx, game); err != nil {
			return err
		}

	case entities.GameStatusActive:
		// Mark as disconnected and start reconnection timer
		player.Disconnect()
		if err := s.playerRepo.SavePlayer(ctx, player); err != nil {
			return err
		}

		s.startReconnTimer(playerID, gameID)

	case entities.GameStatusEnded:
		// Game is over, just mark disconnected
		player.Disconnect()
		if err := s.playerRepo.SavePlayer(ctx, player); err != nil {
			return err
		}
	}

	return nil
}

// LeaveGame removes a player from their current game
// This is used when a player is kicked - it removes them without triggering full disconnect logic
func (s *PlayerService) LeaveGame(ctx context.Context, playerID entities.PlayerID) error {
	player, err := s.playerRepo.GetPlayer(ctx, playerID)
	if err != nil {
		return err
	}

	if player.GameID == nil {
		return nil // Player not in a game
	}

	gameID := *player.GameID

	game, err := s.gameRepo.GetGame(ctx, gameID)
	if err != nil {
		return err
	}

	// Only handle leaving during waiting state (kicks only allowed in waiting state)
	if game.Status != entities.GameStatusWaiting {
		return nil
	}

	// Remove player from game
	if err := s.playerRepo.RemovePlayerFromGame(ctx, gameID, playerID); err != nil {
		return err
	}
	if err := s.playerRepo.DeletePlayer(ctx, playerID); err != nil {
		return err
	}

	// Remove from game.Players slice
	game.Players = removePlayerFromSlice(game.Players, playerID)

	if err := s.gameRepo.SaveGame(ctx, game); err != nil {
		return err
	}

	return nil
}

// GetPlayer retrieves a player by ID
func (s *PlayerService) GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error) {
	return s.playerRepo.GetPlayer(ctx, id)
}

// GetGamePlayers retrieves all players in a game
func (s *PlayerService) GetGamePlayers(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error) {
	return s.playerRepo.GetPlayersByGame(ctx, gameID)
}

// CleanupGamePlayers removes all players when a game ends
func (s *PlayerService) CleanupGamePlayers(ctx context.Context, gameID entities.GameID) error {
	// Cancel all reconnection timers for this game's players
	players, err := s.playerRepo.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return err
	}

	for _, player := range players {
		s.cancelReconnTimer(player.ID)
	}

	return s.playerRepo.DeleteGamePlayers(ctx, gameID)
}

// IsPlayerConnected checks if a player has an active WebSocket session
func (s *PlayerService) IsPlayerConnected(playerID entities.PlayerID) bool {
	if s.connChecker == nil {
		return false
	}
	return s.connChecker.IsPlayerConnected(playerID)
}

// startReconnTimer starts a reconnection timer for a player
func (s *PlayerService) startReconnTimer(playerID entities.PlayerID, gameID entities.GameID) {
	s.timerLock.Lock()
	defer s.timerLock.Unlock()

	// Cancel existing timer if any
	if timer, exists := s.reconnTimers[playerID]; exists {
		timer.Stop()
	}

	s.reconnTimers[playerID] = time.AfterFunc(constants.ReconnectionTimeout, func() {
		s.handleReconnTimeout(playerID, gameID)
	})
}

// cancelReconnTimer cancels a player's reconnection timer
func (s *PlayerService) cancelReconnTimer(playerID entities.PlayerID) {
	s.timerLock.Lock()
	defer s.timerLock.Unlock()

	if timer, exists := s.reconnTimers[playerID]; exists {
		timer.Stop()
		delete(s.reconnTimers, playerID)
	}
}

// handleReconnTimeout is called when a player's reconnection window expires
func (s *PlayerService) handleReconnTimeout(playerID entities.PlayerID, gameID entities.GameID) {
	// Use background context with timeout for timer-triggered operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	player, err := s.playerRepo.GetPlayer(ctx, playerID)
	if err != nil {
		return
	}

	// Only mark inactive if still disconnected
	if player.ConnectionState != entities.ConnectionStateDisconnected {
		return
	}

	// Mark as inactive
	player.SetInactive()
	if err := s.playerRepo.SavePlayer(ctx, player); err != nil {
		return
	}

	// Broadcast inactive event to the game
	if s.connChecker != nil {
		event := events.NewInactiveEvent(playerID)
		payload, err := json.Marshal(event)
		if err == nil {
			s.connChecker.BroadcastToGame(gameID, payload)
		}
	}

	// Cleanup timer
	s.timerLock.Lock()
	delete(s.reconnTimers, playerID)
	s.timerLock.Unlock()
}

// removePlayerFromSlice removes a player ID from a slice
func removePlayerFromSlice(players []entities.PlayerID, playerID entities.PlayerID) []entities.PlayerID {
	for i, p := range players {
		if p == playerID {
			return append(players[:i], players[i+1:]...)
		}
	}
	return players
}

// selectRandomHost selects a random player from the slice to be the new host
func selectRandomHost(players []entities.PlayerID) entities.PlayerID {
	if len(players) == 0 {
		return ""
	}
	if len(players) == 1 {
		return players[0]
	}
	return players[rand.Intn(len(players))]
}
