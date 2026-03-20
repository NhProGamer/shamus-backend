package prompts

import (
	"encoding/json"
	"shamus-backend/internal/domain/entities"
)

// PromptResponse is the structure sent by the client to respond to a prompt
type PromptResponse struct {
	// Channel must be "response"
	Channel entities.Channel `json:"channel"`

	// PromptID is the ID of the prompt being responded to
	PromptID entities.PromptID `json:"promptId"`

	// Response contains the actual response data (varies by prompt type)
	Response json.RawMessage `json:"response"`
}

// SelectPlayerResponse is the response for PromptSelectPlayer
type SelectPlayerResponse struct {
	// PlayerID is the selected player (nil if skipped)
	PlayerID *entities.PlayerID `json:"playerId,omitempty"`

	// Skipped indicates the player chose to skip (if allowed)
	Skipped bool `json:"skipped,omitempty"`
}

// SelectOptionResponse is the response for PromptSelectOption
type SelectOptionResponse struct {
	// OptionID is the ID of the selected option
	OptionID string `json:"optionId,omitempty"`

	// Skipped indicates the player chose to skip (if allowed)
	Skipped bool `json:"skipped,omitempty"`
}

// VoteResponse is the response for PromptVote
type VoteResponse struct {
	// TargetID is the player being voted for (nil if abstaining)
	TargetID *entities.PlayerID `json:"targetId,omitempty"`

	// Abstained indicates the player chose to abstain
	Abstained bool `json:"abstained,omitempty"`
}

// ConfirmResponse is the response for PromptConfirm
type ConfirmResponse struct {
	// Confirmed is true if the player confirmed, false if cancelled
	Confirmed bool `json:"confirmed"`
}

// WitchPotionResponse is the response for witch's potion choice
type WitchPotionResponse struct {
	// OptionID is the chosen option: "heal", "poison", or "skip"
	OptionID string `json:"optionId"`

	// TargetID is the target for poison (required if optionId is "poison")
	TargetID *entities.PlayerID `json:"targetId,omitempty"`
}

// ParseSelectPlayerResponse parses a raw response into SelectPlayerResponse
func ParseSelectPlayerResponse(data json.RawMessage) (*SelectPlayerResponse, error) {
	var resp SelectPlayerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseSelectOptionResponse parses a raw response into SelectOptionResponse
func ParseSelectOptionResponse(data json.RawMessage) (*SelectOptionResponse, error) {
	var resp SelectOptionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseVoteResponse parses a raw response into VoteResponse
func ParseVoteResponse(data json.RawMessage) (*VoteResponse, error) {
	var resp VoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseConfirmResponse parses a raw response into ConfirmResponse
func ParseConfirmResponse(data json.RawMessage) (*ConfirmResponse, error) {
	var resp ConfirmResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseWitchPotionResponse parses a raw response into WitchPotionResponse
func ParseWitchPotionResponse(data json.RawMessage) (*WitchPotionResponse, error) {
	var resp WitchPotionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
