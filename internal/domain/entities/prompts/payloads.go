package prompts

import "shamus-backend/internal/domain/entities"

// SelectPlayerPayload is the payload for PromptSelectPlayer
// Used for: seer vision, witch poison target, etc.
type SelectPlayerPayload struct {
	// Message is the prompt message displayed to the player
	Message string `json:"message"`

	// EligiblePlayers is the list of players that can be selected
	EligiblePlayers []entities.PlayerID `json:"eligiblePlayers"`

	// PlayersInfo provides additional info about each eligible player (optional)
	PlayersInfo map[entities.PlayerID]PlayerInfo `json:"playersInfo,omitempty"`
}

// PlayerInfo contains display information about a player
type PlayerInfo struct {
	Username string `json:"username"`
	IsAlive  bool   `json:"isAlive"`
}

// SelectOptionPayload is the payload for PromptSelectOption
// Used for: witch potion choice (heal/poison/skip), etc.
type SelectOptionPayload struct {
	// Message is the prompt message displayed to the player
	Message string `json:"message"`

	// Options is the list of available options
	Options []Option `json:"options"`
}

// Option represents a selectable option
type Option struct {
	// ID is the unique identifier for this option
	ID string `json:"id"`

	// Label is the display text for the option
	Label string `json:"label"`

	// Description provides additional context (optional)
	Description string `json:"description,omitempty"`

	// Disabled indicates if the option cannot be selected
	Disabled bool `json:"disabled,omitempty"`

	// DisabledReason explains why the option is disabled (if Disabled is true)
	DisabledReason string `json:"disabledReason,omitempty"`
}

// VotePayload is the payload for PromptVote
// Used for: village vote, werewolf vote
type VotePayload struct {
	// Message is the prompt message displayed to the player
	Message string `json:"message"`

	// EligibleTargets is the list of players that can be voted for
	EligibleTargets []entities.PlayerID `json:"eligibleTargets"`

	// CanAbstain indicates if the player can choose not to vote
	CanAbstain bool `json:"canAbstain"`

	// CurrentVotes shows the current votes in a group vote (real-time update)
	// Map of voterID -> targetID (nil if abstained)
	CurrentVotes map[entities.PlayerID]*entities.PlayerID `json:"currentVotes,omitempty"`

	// PlayersInfo provides username info for display
	PlayersInfo map[entities.PlayerID]PlayerInfo `json:"playersInfo,omitempty"`
}

// ConfirmPayload is the payload for PromptConfirm
// Used for: simple yes/no confirmations
type ConfirmPayload struct {
	// Message is the question to confirm
	Message string `json:"message"`

	// ConfirmText is the text for the confirm button (default: "Yes")
	ConfirmText string `json:"confirmText,omitempty"`

	// CancelText is the text for the cancel button (default: "No")
	CancelText string `json:"cancelText,omitempty"`
}

// WitchPotionPayload is a specialized payload for the witch's turn
// Combines information about victim and available potions
type WitchPotionPayload struct {
	// Message is the prompt message
	Message string `json:"message"`

	// VictimID is the player killed by werewolves (nil if no victim)
	VictimID *entities.PlayerID `json:"victimId,omitempty"`

	// VictimUsername is the username of the victim (for display)
	VictimUsername string `json:"victimUsername,omitempty"`

	// Options for the witch
	Options []Option `json:"options"`
}

// NewWitchPotionPayload creates a payload for the witch's potion choice
func NewWitchPotionPayload(victimID *entities.PlayerID, victimUsername string, canHeal, canPoison bool) *WitchPotionPayload {
	options := make([]Option, 0, 3)

	// Heal option
	healOption := Option{
		ID:    "heal",
		Label: "Soigner",
	}
	if victimID != nil {
		healOption.Description = "Sauver " + victimUsername + " de l'attaque des loups"
	} else {
		healOption.Description = "Personne n'a été attaqué cette nuit"
		healOption.Disabled = true
		healOption.DisabledReason = "Aucune victime à soigner"
	}
	if !canHeal {
		healOption.Disabled = true
		healOption.DisabledReason = "Potion de soin déjà utilisée"
	}
	options = append(options, healOption)

	// Poison option
	poisonOption := Option{
		ID:          "poison",
		Label:       "Empoisonner",
		Description: "Tuer un joueur de votre choix",
	}
	if !canPoison {
		poisonOption.Disabled = true
		poisonOption.DisabledReason = "Potion de poison déjà utilisée"
	}
	options = append(options, poisonOption)

	// Skip option
	options = append(options, Option{
		ID:          "skip",
		Label:       "Passer",
		Description: "Ne rien faire cette nuit",
	})

	message := "Que souhaitez-vous faire ?"
	if victimID != nil {
		message = victimUsername + " a été attaqué par les loups. Que souhaitez-vous faire ?"
	}

	return &WitchPotionPayload{
		Message:        message,
		VictimID:       victimID,
		VictimUsername: victimUsername,
		Options:        options,
	}
}
