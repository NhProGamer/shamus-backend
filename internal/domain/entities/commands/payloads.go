package commands

import (
	"encoding/json"
	"shamus-backend/internal/domain/entities"
)

// ChatChannel represents the different chat channels
type ChatChannel string

const (
	ChatChannelVillage  ChatChannel = "village"  // All alive players during day
	ChatChannelWerewolf ChatChannel = "werewolf" // Werewolves only during night
	ChatChannelDead     ChatChannel = "dead"     // Dead players only
)

// SendChatPayload is the payload for CmdSendChat
type SendChatPayload struct {
	Message string      `json:"message"`
	Channel ChatChannel `json:"channel"`
}

// UpdateSettingsPayload is the payload for CmdUpdateSettings
type UpdateSettingsPayload struct {
	Roles map[entities.RoleType]int `json:"roles"`
}

// StartGamePayload is the payload for CmdStartGame (empty, but defined for consistency)
type StartGamePayload struct{}

// LeaveGamePayload is the payload for CmdLeaveGame (empty, but defined for consistency)
type LeaveGamePayload struct{}

// KickPlayerPayload is the payload for CmdKickPlayer
type KickPlayerPayload struct {
	PlayerID entities.PlayerID `json:"playerId"`
	Reason   string            `json:"reason,omitempty"`
}

// ParseSendChatPayload parses a raw payload into SendChatPayload
func ParseSendChatPayload(data json.RawMessage) (*SendChatPayload, error) {
	var payload SendChatPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// ParseUpdateSettingsPayload parses a raw payload into UpdateSettingsPayload
func ParseUpdateSettingsPayload(data json.RawMessage) (*UpdateSettingsPayload, error) {
	var payload UpdateSettingsPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// ParseKickPlayerPayload parses a raw payload into KickPlayerPayload
func ParseKickPlayerPayload(data json.RawMessage) (*KickPlayerPayload, error) {
	var payload KickPlayerPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
