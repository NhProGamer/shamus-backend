package entities

type EventType string

type EventChannel string

const (
	EventChannelGameEvent EventChannel = "game_event"
	EventChannelConnexion EventChannel = "conn_event"
	EventChannelSettings  EventChannel = "settings_event"
	EventChannelTimer     EventChannel = "timer_event"
)

type Event struct {
	Channel EventChannel `json:"channel"`
	Type    EventType    `json:"type"`
	Data    interface{}  `json:"data"`
}

func (e Event) GetChannel() EventChannel {
	return e.Channel
}

func (e Event) GetType() EventType {
	return e.Type
}

func (e Event) GetData() interface{} {
	return e.Data
}
