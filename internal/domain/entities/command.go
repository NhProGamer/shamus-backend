package entities

import "encoding/json"

// CommandType represents the type of command sent by the client
type CommandType string

const (
	// CmdSendChat sends a chat message
	CmdSendChat CommandType = "send_chat"

	// CmdUpdateSettings updates game settings (host only)
	CmdUpdateSettings CommandType = "update_settings"

	// CmdStartGame starts the game (host only)
	CmdStartGame CommandType = "start_game"

	// CmdLeaveGame leaves the current game
	CmdLeaveGame CommandType = "leave_game"

	// CmdKickPlayer kicks a player from the game (host only)
	CmdKickPlayer CommandType = "kick_player"
)

// Command represents a client-to-server command message
type Command struct {
	Channel Channel         `json:"channel"`
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// CommandMessage is an alias for Command (for clarity in handler code)
type CommandMessage = Command

// ParseCommand parses a raw message into a Command
func ParseCommand(data []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}
