package entities

type VoteType string
type VoteStatus string

type Vote struct {
	ID     string
	Type   VoteType
	Status VoteStatus

	// TODO: Add vote timing concepts (start time, end time, duration, etc.)
	EligibleVoters  []PlayerID
	EligibleTargets []PlayerID
	Ballots         map[PlayerID]*PlayerID
	AllowAbstain    bool

	Result *VoteResult
}

type VoteResult struct {
	Target      *PlayerID
	Counts      map[PlayerID]int
	IsTie       bool
	TiedPlayers []PlayerID
}
