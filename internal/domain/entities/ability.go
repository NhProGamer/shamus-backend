package entities

type Ability interface {
	GetName() string
	GetDescription() string
	CanUse(game *Game, player *Player) bool
	GetConsumptions() *uint8
	Consume()
	Execute(game *Game, player *Player, target *PlayerID, data map[string]interface{}) error
}
