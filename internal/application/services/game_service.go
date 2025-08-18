package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

type gameService struct {
	gameRepo       ports.GameRepository
	playerRepo     ports.PlayerRepository
	settingsRepo   ports.GameSettingsRepository
	votesRepo      ports.GameVotesRepository
	eventService   ports.EventService
	abilityService ports.AbilityService
	voteService    ports.VoteService
}

func NewGameService(
	gr ports.GameRepository,
	pr ports.PlayerRepository,
	sr ports.GameSettingsRepository,
	vr ports.GameVotesRepository,
	es ports.EventService,
	as ports.AbilityService,
	vs ports.VoteService,
) ports.GameService {
	return &gameService{
		gameRepo:       gr,
		playerRepo:     pr,
		settingsRepo:   sr,
		votesRepo:      vr,
		eventService:   es,
		abilityService: as,
		voteService:    vs,
	}
}

func (s *gameService) NextPhase(gameID entities.GameID) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	switch game.Phase {
	case entities.PhaseDay:
		game.Phase = entities.PhaseVote
	case entities.PhaseNight:
		game.Phase = entities.PhaseDay
		game.Day++
	case entities.PhaseVote:
		game.Phase = entities.PhaseNight
	}

	return s.gameRepo.UpdateGame(game)
}

func (s *gameService) NextStep(gameID entities.GameID) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	switch game.Phase {
	case entities.PhaseDay:
		//Envoyer qui est mort durant la nuit
		return s.NextPhase(gameID)

	case entities.PhaseVote:
		victim, err := s.voteService.ComputeVotes(gameID)
		if err != nil {
			return err
		}

		if victim != nil {
			// Marquer le joueur comme mort
			player, err := s.playerRepo.GetPlayerByID(*victim)
			if err != nil {
				return err
			}
			player.Status = entities.PlayerStatusDead
			// Envoyer l'événement d'exécution
			event := events.ExecutionEvent{
				VictimID: victim,
			}
			s.eventService.SendEventToGame(event, gameID)
		}

		// Réinitialiser les votes pour le prochain tour
		if err := s.voteService.CleanVotes(gameID); err != nil {
			return err
		}

		// Passer à la phase suivante
		return s.NextPhase(gameID)

	case entities.PhaseNight:
		// Chaque joueur ayant un role actif peut utiliser son pouvoir en suivant l'odre de priorité
		// 0 équivaut a ne rien pouvoir faire la nuit
		return s.NextPhase(gameID)

	default:
		return s.NextPhase(gameID)
	}
}

func (s *gameService) IsGameEnded(gameID entities.GameID) (bool, error) {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return false, err
	}

	// Compter les joueurs par clan
	clanCount := make(map[entities.Clan]int)
	for _, player := range game.Players {
		if player.Status == entities.PlayerStatusAlive {
			for _, clan := range player.Role.GetClans() {
				clanCount[clan]++
			}
		}
	}

	// Conditions de victoire
	villagers := clanCount[entities.ClanVillager]
	werewolves := clanCount[entities.ClanWerewolf]

	switch {
	case werewolves == 0:
		// Victoire des villageois
		game.Status = entities.GameStatusEnded
		return true, s.gameRepo.UpdateGame(game)
	case werewolves >= villagers:
		// Victoire des loups-garous
		game.Status = entities.GameStatusEnded
		return true, s.gameRepo.UpdateGame(game)
	}

	return false, nil
}
