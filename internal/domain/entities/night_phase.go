package entities

// NightPhase represents the current phase of the night
type NightPhase string

const (
	NightPhaseSeer     NightPhase = "seer"
	NightPhaseWerewolf NightPhase = "werewolf"
	NightPhaseWitch    NightPhase = "witch"
	NightPhaseEnd      NightPhase = "end"
)

// NightPhaseOrder defines the order of night phases (by role priority)
// Seer acts first (priority 10), then Werewolves (priority 9), then Witch (priority 8)
var NightPhaseOrder = []NightPhase{
	NightPhaseSeer,
	NightPhaseWerewolf,
	NightPhaseWitch,
}
