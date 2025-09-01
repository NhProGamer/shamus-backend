package events

import "shamus-backend/internal/domain/entities"

const EventTypeGameData entities.EventType = "settings"

type PlayersDetailsData struct {
	ID             entities.PlayerID        `json:"id"`
	Alive          bool                     `json:"alive"`
	Role           *entities.RoleType       `json:"role,omitempty"`
	Target         entities.PlayerID        `json:"target,omitempty"`
	ConnexionState entities.ConnectionState `json:"connexion_state"`
}

type GameSettingsEventData struct {
	ID       entities.GameID       `json:"id"`
	Status   entities.GameStatus   `json:"status"`
	Phase    entities.GamePhase    `json:"phase"`
	Day      int                   `json:"day"`
	Players  []PlayersDetailsData  `json:"players"`
	Host     entities.PlayerID     `json:"host"`
	Settings entities.GameSettings `json:"settings"`
}

func NewGameDataEvent(data GameSettingsEventData) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelSettings,
		Type:    EventTypeGameData,
		Data:    data,
	}
}
