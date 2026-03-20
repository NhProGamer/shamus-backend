package entities

import (
	"encoding/json"
	"time"
)

// PromptID is a unique identifier for a prompt
type PromptID string

// GroupID is a unique identifier for a group of related prompts (e.g., werewolf vote)
type GroupID string

// PromptType represents the type of interaction requested from the player
type PromptType string

const (
	// PromptSelectPlayer asks the player to select one player from a list
	PromptSelectPlayer PromptType = "select_player"

	// PromptSelectOption asks the player to select one option from a list
	PromptSelectOption PromptType = "select_option"

	// PromptVote asks the player to vote (can abstain)
	PromptVote PromptType = "vote"

	// PromptConfirm asks for a yes/no confirmation
	PromptConfirm PromptType = "confirm"
)

// PromptStatus represents the current status of a prompt
type PromptStatus string

const (
	PromptStatusPending   PromptStatus = "pending"   // Waiting for response
	PromptStatusAnswered  PromptStatus = "answered"  // Response received
	PromptStatusExpired   PromptStatus = "expired"   // Timeout reached
	PromptStatusCancelled PromptStatus = "cancelled" // Manually cancelled
)

// Prompt represents a request for player interaction with timeout
type Prompt struct {
	ID       PromptID        `json:"id"`
	Type     PromptType      `json:"type"`
	Context  string          `json:"context"` // e.g., "seer_vision", "werewolf_vote", "village_vote"
	GameID   GameID          `json:"gameId"`
	PlayerID PlayerID        `json:"playerId"`
	Payload  json.RawMessage `json:"payload"`
	Status   PromptStatus    `json:"status"`

	// Timing
	Timeout     time.Duration `json:"-"`         // Internal use
	ExpiresAt   time.Time     `json:"expiresAt"`
	TimeoutSecs int           `json:"timeout"`   // For client display

	// Options
	AllowChange bool `json:"allowChange"` // Can re-submit response (for group votes)
	CanSkip     bool `json:"canSkip"`     // Can skip without responding

	// Group voting (for werewolf/village votes)
	GroupID *GroupID `json:"groupId,omitempty"`

	// Response storage
	Response json.RawMessage `json:"-"` // Stored response (internal)
}

// PromptMessage is the message sent to the client
type PromptMessage struct {
	Channel     Channel         `json:"channel"`
	Type        PromptType      `json:"type"`
	ID          PromptID        `json:"id"`
	Context     string          `json:"context"`
	Payload     json.RawMessage `json:"payload"`
	ExpiresAt   time.Time       `json:"expiresAt"`
	TimeoutSecs int             `json:"timeout"`
	AllowChange bool            `json:"allowChange"`
	CanSkip     bool            `json:"canSkip"`
	GroupID     *GroupID        `json:"groupId,omitempty"`
}

// NewPrompt creates a new prompt
func NewPrompt(
	id PromptID,
	promptType PromptType,
	context string,
	gameID GameID,
	playerID PlayerID,
	payload interface{},
	timeout time.Duration,
	allowChange bool,
	canSkip bool,
	groupID *GroupID,
) (*Prompt, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Prompt{
		ID:          id,
		Type:        promptType,
		Context:     context,
		GameID:      gameID,
		PlayerID:    playerID,
		Payload:     payloadBytes,
		Status:      PromptStatusPending,
		Timeout:     timeout,
		ExpiresAt:   now.Add(timeout),
		TimeoutSecs: int(timeout.Seconds()),
		AllowChange: allowChange,
		CanSkip:     canSkip,
		GroupID:     groupID,
	}, nil
}

// ToMessage converts the prompt to a client message
func (p *Prompt) ToMessage() *PromptMessage {
	return &PromptMessage{
		Channel:     ChannelPrompt,
		Type:        p.Type,
		ID:          p.ID,
		Context:     p.Context,
		Payload:     p.Payload,
		ExpiresAt:   p.ExpiresAt,
		TimeoutSecs: p.TimeoutSecs,
		AllowChange: p.AllowChange,
		CanSkip:     p.CanSkip,
		GroupID:     p.GroupID,
	}
}

// IsExpired checks if the prompt has exceeded its timeout
func (p *Prompt) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// CanRespond checks if the prompt can still receive a response
func (p *Prompt) CanRespond() bool {
	if p.IsExpired() {
		return false
	}
	// If AllowChange is true, can respond multiple times while pending
	if p.AllowChange {
		return p.Status == PromptStatusPending || p.Status == PromptStatusAnswered
	}
	return p.Status == PromptStatusPending
}

// MarkAnswered marks the prompt as answered with the given response
func (p *Prompt) MarkAnswered(response json.RawMessage) {
	p.Status = PromptStatusAnswered
	p.Response = response
}

// MarkExpired marks the prompt as expired
func (p *Prompt) MarkExpired() {
	p.Status = PromptStatusExpired
}

// MarkCancelled marks the prompt as cancelled
func (p *Prompt) MarkCancelled() {
	p.Status = PromptStatusCancelled
}

// GetRemainingTime returns the time remaining before expiry
func (p *Prompt) GetRemainingTime() time.Duration {
	remaining := time.Until(p.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}
