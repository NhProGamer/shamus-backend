package events

import "shamus-backend/internal/domain/entities"

const EventTypeVote entities.EventType = "vote"

type VoteEvent struct {
}

func (e VoteEvent) GetType() entities.EventType {
	return EventTypeVote
}

func (e VoteEvent) GetData() interface{} {
	return e
}
