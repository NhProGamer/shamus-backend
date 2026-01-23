package entities

type VoteType string
type VoteStatus string

const (
	VoteTypeVillage  VoteType = "village"
	VoteTypeWerewolf VoteType = "werewolf"
)

const (
	VoteStatusPending  VoteStatus = "pending"
	VoteStatusActive   VoteStatus = "active"
	VoteStatusResolved VoteStatus = "resolved"
)

type Vote struct {
	ID     string
	Type   VoteType
	Status VoteStatus

	EligibleVoters  []PlayerID
	EligibleTargets []PlayerID
	Ballots         map[PlayerID]*PlayerID
	AllowAbstain    bool

	Result *VoteResult
}

// NewVote creates a new Vote with the given parameters
func NewVote(id string, voteType VoteType, voters, targets []PlayerID, allowAbstain bool) *Vote {
	return &Vote{
		ID:              id,
		Type:            voteType,
		Status:          VoteStatusActive,
		EligibleVoters:  voters,
		EligibleTargets: targets,
		Ballots:         make(map[PlayerID]*PlayerID),
		AllowAbstain:    allowAbstain,
		Result:          nil,
	}
}

// CastBallot records a vote from a voter to a target
func (v *Vote) CastBallot(voterID PlayerID, targetID *PlayerID) bool {
	// Check if voter is eligible
	isEligible := false
	for _, id := range v.EligibleVoters {
		if id == voterID {
			isEligible = true
			break
		}
	}
	if !isEligible {
		return false
	}

	// Check if target is eligible (if not abstaining)
	if targetID != nil {
		targetEligible := false
		for _, id := range v.EligibleTargets {
			if id == *targetID {
				targetEligible = true
				break
			}
		}
		if !targetEligible {
			return false
		}
	} else if !v.AllowAbstain {
		return false
	}

	v.Ballots[voterID] = targetID
	return true
}

// HasEveryoneVoted checks if all eligible voters have cast their ballot
func (v *Vote) HasEveryoneVoted() bool {
	// No eligible voters means voting cannot complete normally
	if len(v.EligibleVoters) == 0 {
		return false
	}
	return len(v.Ballots) >= len(v.EligibleVoters)
}

// Resolve computes the result of the vote
// Returns nil target if there's a tie (no elimination)
func (v *Vote) Resolve() *VoteResult {
	counts := make(map[PlayerID]int)

	// Count votes
	for _, target := range v.Ballots {
		if target != nil {
			counts[*target]++
		}
	}

	// Find the maximum vote count
	maxVotes := 0
	var tiedPlayers []PlayerID

	for id, count := range counts {
		if count > maxVotes {
			maxVotes = count
			tiedPlayers = []PlayerID{id}
		} else if count == maxVotes && maxVotes > 0 {
			tiedPlayers = append(tiedPlayers, id)
		}
	}

	result := &VoteResult{
		Counts:      counts,
		IsTie:       len(tiedPlayers) > 1,
		TiedPlayers: tiedPlayers,
	}

	// Only set target if there's no tie
	if len(tiedPlayers) == 1 {
		result.Target = &tiedPlayers[0]
	}

	v.Result = result
	v.Status = VoteStatusResolved

	return result
}

type VoteResult struct {
	Target      *PlayerID
	Counts      map[PlayerID]int
	IsTie       bool
	TiedPlayers []PlayerID
}
