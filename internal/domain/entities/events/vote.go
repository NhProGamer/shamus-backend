package events

import "shamus-backend/internal/domain/entities"

const EventTypeVote entities.EventType = "vote"

type VoteEventType string

const (
	StartVote  VoteEventType = "start"
	EndVote    VoteEventType = "end"
	PlayerVote VoteEventType = "player"
)

type VoteEventData struct {
	Type   VoteEventType      `json:"type"`
	Player *entities.PlayerID `json:"player,omitempty"`
	Target *entities.PlayerID `json:"target,omitempty"`
}

func NewVoteEvent(voteType VoteEventType, player *entities.PlayerID, target *entities.PlayerID) entities.Event {
	return entities.Event{
		Type: EventTypeVote,
		Data: VoteEventData{
			Type:   voteType,
			Player: player,
			Target: target,
		},
	}

}
