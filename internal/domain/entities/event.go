package entities

type EventType string

type Event interface {
	GetType() EventType
	GetData() interface{}
}
