package entities

// Ability represents a role's special power.
// The execution logic is handled by GameEngine which has access to
// repositories and services. Abilities are value objects that track
// state (like consumptions) but don't perform side effects themselves.
type Ability interface {
	GetName() string
	GetDescription() string
	CanUse(game *Game, player *Player) bool
	GetConsumptions() *uint8
	// TryConsume attempts to consume one use of the ability.
	// Returns true if successful, false if no consumptions remaining.
	// For unlimited abilities, always returns true.
	TryConsume() bool
}
