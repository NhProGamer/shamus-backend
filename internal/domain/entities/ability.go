package entities

type Ability interface {
	GetName() string
	GetDescription() string
	CanUse(game *Game, player *Player) bool
	Execute(game *Game, player *Player, target *PlayerID, data map[string]interface{}) error
}
