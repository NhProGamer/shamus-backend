package events

import "shamus-backend/internal/domain/entities"

const EventTypeWin entities.EventType = "win"

type WinEventData struct {
	WinningClan entities.Clan       `json:"winningClan"`
	Winners     []entities.PlayerID `json:"winners"`
}

func NewWinEvent(winningClan entities.Clan, winners []entities.PlayerID) entities.Event[WinEventData] {
	return entities.Event[WinEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeWin,
		Data: WinEventData{
			WinningClan: winningClan,
			Winners:     winners,
		},
	}
}
