package errors

import "errors"

// Erreurs typiques liées aux capacités dans le jeu du Loup-Garou
var (
	// Capacité épuisée (ex: potion déjà utilisée)
	ErrNoUsagesLeft = errors.New("no usages left for this ability")

	// Capacité utilisée dans une phase invalide (ex: essayer de voter en journée)
	ErrInvalidPhase = errors.New("ability cannot be used in this phase")

	// Cible invalide (ex: soigner un joueur mort, s’auto-cibler si interdit, etc.)
	ErrInvalidTarget = errors.New("invalid target for this ability")

	// Joueur déjà mort, donc il ne peut pas utiliser la capacité
	ErrPlayerDead = errors.New("dead players cannot use abilities")

	// Capacité déjà utilisée ce tour (si usage limité par phase)
	ErrAlreadyUsed = errors.New("ability already used in this turn")

	// Données manquantes pour exécuter la capacité (ex: pas de cible, pas d’arguments)
	ErrMissingData = errors.New("missing data for ability execution")

	// Action interdite par les règles du rôle (ex: la Voyante veut voir deux personnes à la fois)
	ErrForbiddenAction = errors.New("forbidden action for this role/ability")
)
