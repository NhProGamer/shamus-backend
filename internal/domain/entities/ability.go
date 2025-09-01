package entities

type Ability interface {
	GetName() string
	GetDescription() string
	CanUse(game *Game, player *SafePlayer) bool
	GetConsumptions() *uint8
	Consume()
	Execute(game *Game, player *SafePlayer, target *PlayerID, data map[string]interface{}) error
}
