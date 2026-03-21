package services

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/notifications"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
)

// NotificationService handles sending notifications to players
// It provides typed methods for common notifications and abstracts the serialization
type NotificationService struct {
	broadcaster ports.Broadcaster
	sender      ports.PlayerSender
	playerRepo  ports.PlayerRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	broadcaster ports.Broadcaster,
	sender ports.PlayerSender,
	playerRepo ports.PlayerRepository,
) *NotificationService {
	return &NotificationService{
		broadcaster: broadcaster,
		sender:      sender,
		playerRepo:  playerRepo,
	}
}

// --- Low-level methods ---

// NotifyPlayer sends a notification to a specific player
func (s *NotificationService) NotifyPlayer(playerID entities.PlayerID, notifType entities.NotificationType, payload interface{}) error {
	notif, err := entities.NewNotification(notifType, payload)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to create notification")
		return err
	}

	data, err := json.Marshal(notif)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to marshal notification")
		return err
	}

	return s.sender.SendToPlayer(playerID, data)
}

// NotifyPlayers sends a notification to multiple specific players
func (s *NotificationService) NotifyPlayers(playerIDs []entities.PlayerID, notifType entities.NotificationType, payload interface{}) {
	notif, err := entities.NewNotification(notifType, payload)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to create notification")
		return
	}

	data, err := json.Marshal(notif)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to marshal notification")
		return
	}

	for _, playerID := range playerIDs {
		if err := s.sender.SendToPlayer(playerID, data); err != nil {
			logger.Get().Warn().Err(err).Str("playerID", string(playerID)).Msg("Failed to send notification to player")
		}
	}
}

// NotifyAll sends a notification to all players in a game
func (s *NotificationService) NotifyAll(gameID entities.GameID, notifType entities.NotificationType, payload interface{}) error {
	notif, err := entities.NewNotification(notifType, payload)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to create notification")
		return err
	}

	data, err := json.Marshal(notif)
	if err != nil {
		logger.Get().Error().Err(err).Str("type", string(notifType)).Msg("Failed to marshal notification")
		return err
	}

	return s.broadcaster.BroadcastToGame(gameID, data)
}

// NotifyRole sends a notification to all players with a specific role
func (s *NotificationService) NotifyRole(gameID entities.GameID, roleType entities.RoleType, notifType entities.NotificationType, payload interface{}) error {
	players, err := s.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	var targetPlayers []entities.PlayerID
	for _, p := range players {
		if p.Role != nil && p.Role.GetType() == roleType && p.IsAlive {
			targetPlayers = append(targetPlayers, p.ID)
		}
	}

	s.NotifyPlayers(targetPlayers, notifType, payload)
	return nil
}

// --- Typed helper methods ---

// NotifyPhaseChanged notifies all players of a phase change
func (s *NotificationService) NotifyPhaseChanged(gameID entities.GameID, phase entities.GamePhase, day int, subPhase string) error {
	return s.NotifyAll(gameID, entities.NotifPhaseChanged, notifications.PhaseChangedPayload{
		Phase:    phase,
		Day:      day,
		SubPhase: subPhase,
	})
}

// NotifyPlayerJoined notifies all players that a new player joined
func (s *NotificationService) NotifyPlayerJoined(gameID entities.GameID, playerID entities.PlayerID, username string) error {
	return s.NotifyAll(gameID, entities.NotifPlayerJoined, notifications.PlayerJoinedPayload{
		PlayerID: playerID,
		Username: username,
	})
}

// NotifyPlayerLeft notifies all players that a player left
func (s *NotificationService) NotifyPlayerLeft(gameID entities.GameID, playerID entities.PlayerID, username string, reason string) error {
	return s.NotifyAll(gameID, entities.NotifPlayerLeft, notifications.PlayerLeftPayload{
		PlayerID: playerID,
		Username: username,
		Reason:   reason,
	})
}

// NotifyPlayerDied notifies all players of a death
func (s *NotificationService) NotifyPlayerDied(gameID entities.GameID, playerID entities.PlayerID, username string, role entities.RoleType, cause string) error {
	return s.NotifyAll(gameID, entities.NotifPlayerDied, notifications.PlayerDiedPayload{
		PlayerID: playerID,
		Username: username,
		Role:     role,
		Cause:    cause,
	})
}

// NotifyHostChanged notifies all players that the host has changed
func (s *NotificationService) NotifyHostChanged(gameID entities.GameID, newHostID entities.PlayerID, newHostUsername string) error {
	return s.NotifyAll(gameID, entities.NotifHostChanged, notifications.HostChangedPayload{
		NewHostID:       newHostID,
		NewHostUsername: newHostUsername,
	})
}

// NotifyPlayerInactive notifies all players that a player became inactive
func (s *NotificationService) NotifyPlayerInactive(gameID entities.GameID, playerID entities.PlayerID, username string) error {
	return s.NotifyAll(gameID, entities.NotifPlayerInactive, notifications.PlayerInactivePayload{
		PlayerID: playerID,
		Username: username,
	})
}

// NotifyGameStarted notifies all players that the game started
func (s *NotificationService) NotifyGameStarted(gameID entities.GameID, day int) error {
	return s.NotifyAll(gameID, entities.NotifGameStarted, notifications.GameStartedPayload{
		Day: day,
	})
}

// NotifyGameEnded notifies all players that the game ended
func (s *NotificationService) NotifyGameEnded(gameID entities.GameID, winningClan entities.Clan, winners []entities.PlayerID) error {
	return s.NotifyAll(gameID, entities.NotifGameEnded, notifications.GameEndedPayload{
		WinningClan: winningClan,
		Winners:     winners,
	})
}

// NotifyRoleReveal sends a player their assigned role (private)
func (s *NotificationService) NotifyRoleReveal(playerID entities.PlayerID, role entities.RoleType, roleName, description string) error {
	return s.NotifyPlayer(playerID, entities.NotifRoleReveal, notifications.RoleRevealPayload{
		Role:        role,
		RoleName:    roleName,
		Description: description,
	})
}

// NotifyTimerStarted notifies all players that a timer has started
func (s *NotificationService) NotifyTimerStarted(gameID entities.GameID, phase entities.GamePhase, subPhase string, durationSec int) error {
	return s.NotifyAll(gameID, entities.NotifTimerStarted, notifications.TimerStartedPayload{
		Phase:       phase,
		SubPhase:    subPhase,
		DurationSec: durationSec,
	})
}

// NotifyTimerTick notifies all players of timer progress
func (s *NotificationService) NotifyTimerTick(gameID entities.GameID, phase entities.GamePhase, subPhase string, remainingSec int) error {
	return s.NotifyAll(gameID, entities.NotifTimerTick, notifications.TimerTickPayload{
		Phase:        phase,
		SubPhase:     subPhase,
		RemainingSec: remainingSec,
	})
}

// NotifyTimerExpired notifies all players that a timer expired
func (s *NotificationService) NotifyTimerExpired(gameID entities.GameID, phase entities.GamePhase, subPhase string) error {
	return s.NotifyAll(gameID, entities.NotifTimerExpired, notifications.TimerExpiredPayload{
		Phase:    phase,
		SubPhase: subPhase,
	})
}

// NotifySeerResult sends the seer their vision result (private)
func (s *NotificationService) NotifySeerResult(seerID entities.PlayerID, targetID entities.PlayerID, targetUsername string, role entities.RoleType) error {
	return s.NotifyPlayer(seerID, entities.NotifSeerResult, notifications.SeerResultPayload{
		TargetID:       targetID,
		TargetUsername: targetUsername,
		Role:           role,
	})
}

// NotifyVoteStarted notifies relevant players that a vote has started
func (s *NotificationService) NotifyVoteStarted(gameID entities.GameID, voteType string, voters, targets []entities.PlayerID, durationSec int) error {
	return s.NotifyAll(gameID, entities.NotifVoteStarted, notifications.VoteStartedPayload{
		VoteType:        voteType,
		EligibleVoters:  voters,
		EligibleTargets: targets,
		DurationSec:     durationSec,
	})
}

// NotifyVoteUpdate sends real-time vote update to group members
func (s *NotificationService) NotifyVoteUpdate(playerIDs []entities.PlayerID, votes map[entities.PlayerID]*entities.PlayerID, voteCounts map[entities.PlayerID]int, votersRemaining []entities.PlayerID) {
	s.NotifyPlayers(playerIDs, entities.NotifVoteUpdate, notifications.VoteUpdatePayload{
		Votes:           votes,
		VoteCounts:      voteCounts,
		VotersRemaining: votersRemaining,
	})
}

// NotifyVoteResult notifies all players of the vote result
func (s *NotificationService) NotifyVoteResult(gameID entities.GameID, voteType string, result *entities.PlayerID, isTie bool, tiedPlayers []entities.PlayerID, finalVotes map[entities.PlayerID]int, mayorDecided bool) error {
	return s.NotifyAll(gameID, entities.NotifVoteResult, notifications.VoteResultPayload{
		VoteType:     voteType,
		Result:       result,
		IsTie:        isTie,
		TiedPlayers:  tiedPlayers,
		FinalVotes:   finalVotes,
		MayorDecided: mayorDecided,
	})
}

// NotifyMayorTiebreaker notifies the mayor they must break a tie
func (s *NotificationService) NotifyMayorTiebreaker(mayorID entities.PlayerID, tiedPlayers []entities.PlayerID, voteCounts map[entities.PlayerID]int) error {
	return s.NotifyPlayer(mayorID, entities.NotifMayorTiebreaker, notifications.MayorTiebreakerPayload{
		TiedPlayers: tiedPlayers,
		VoteCounts:  voteCounts,
	})
}

// NotifyChatMessage broadcasts a chat message to appropriate recipients
func (s *NotificationService) NotifyChatMessage(recipients []entities.PlayerID, senderID entities.PlayerID, senderName, message, channel string, timestamp int64) {
	s.NotifyPlayers(recipients, entities.NotifChatMessage, notifications.ChatMessagePayload{
		SenderID:   senderID,
		SenderName: senderName,
		Message:    message,
		Channel:    channel,
		Timestamp:  timestamp,
	})
}

// NotifyError sends an error notification to a player
func (s *NotificationService) NotifyError(playerID entities.PlayerID, code, message, action string) error {
	return s.NotifyPlayer(playerID, entities.NotifError, notifications.ErrorPayload{
		Code:    code,
		Message: message,
		Action:  action,
	})
}

// NotifyAck sends an acknowledgement to a player
func (s *NotificationService) NotifyAck(playerID entities.PlayerID, action string, success bool, message string) error {
	return s.NotifyPlayer(playerID, entities.NotifAck, notifications.AckPayload{
		Action:  action,
		Success: success,
		Message: message,
	})
}
