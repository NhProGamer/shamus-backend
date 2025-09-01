package adapters

import (
	"math/rand/v2"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/entities/factories"
	"shamus-backend/internal/domain/ports"
)

type gameService struct {
	gameRepo     ports.GameRepository
	playerRepo   ports.PlayerRepository
	votesRepo    ports.GameVotesRepository
	eventService ports.EventService
	voteService  ports.VoteService
}

func NewGameService(
	gr ports.GameRepository,
	pr ports.PlayerRepository,
	vr ports.GameVotesRepository,
	es ports.EventService,
	vs ports.VoteService,
) ports.GameService {
	return &gameService{
		gameRepo:     gr,
		playerRepo:   pr,
		votesRepo:    vr,
		eventService: es,
		voteService:  vs,
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
		break
	case entities.PhaseNight:
		game.Phase = entities.PhaseDay
		game.Day++
		break
	case entities.PhaseVote:
		game.Phase = entities.PhaseNight
		break
	case entities.PhaseStart:
		game.Phase = entities.PhaseNight
		break
	}

	return nil
}

func (s *gameService) NextStep(gameID entities.GameID) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	switch game.Phase {
	case entities.PhaseStart:
		// Envoyer aux clients que la nuit a commencé
		event := events.NewNightEvent()
		s.eventService.SendEventToGame(event, game.ID)

		// Attribuer les rôles aux joueurs aléatoirement
		rand.Shuffle(len(game.Players), func(i, j int) {
			game.Players[i], game.Players[j] = game.Players[j], game.Players[i]
		})
		actualPlayerInList := 0

		for role, nb := range game.Settings.Roles {
			for i := 0; i < nb; i++ {
				playerID := game.Players[actualPlayerInList]
				actualPlayer, _ := s.playerRepo.GetPlayerByID(playerID)
				pRole := factories.GetNewRole(role)
				actualPlayer.AssignRole(&pRole)

				roleAttributionEvent := events.NewRoleAttributionEvent(role)
				s.eventService.SendEventToPlayer(roleAttributionEvent, playerID)

				actualPlayerInList++
			}
		}
		return s.NextPhase(gameID)

	case entities.PhaseDay:
		// Envoyer aux clients que la journée a commencé, et qui est mort durant la nuit
		event := events.NewDayEvent([]entities.PlayerID{
			// TODO: Ajouter les joueurs morts durant la nuit
		})
		s.eventService.SendEventToGame(event, game.ID)

		isGameEnded, _ := s.IsGameEnded(game.ID)
		if isGameEnded {
			event := events.NewWinEvent([]entities.PlayerID{
				// TODO: Ajouter les joueurs gagnants
			})
			s.eventService.SendEventToGame(event, game.ID)
		}
		return s.NextPhase(gameID)

	case entities.PhaseVote:
		// Envoyer aux clients que le vote commence
		err := s.voteService.NewVote(game.ID)

		// TODO: Implémenter le timer pendant le vote

		// Fermer le vote et tuer celui qui a été le plus voté
		victim, err := s.voteService.CloseVote(game.ID)
		if err != nil {
			return err
		}

		if victim != nil {
			// Marquer le joueur comme mort
			player, err := s.playerRepo.GetPlayerByID(*victim)
			if err != nil {
				return err
			}
			player.Kill()

			// Envoyer qui est mort durant le vote
			event := events.NewDeathEvent(nil, player.ID())
			s.eventService.SendEventToGame(event, gameID)
		} else {
			//TODO: Le village n'a pas réussi à s'accorder sur une victime
		}

		// Passer à la phase suivante
		return s.NextPhase(gameID)

	case entities.PhaseNight:
		// Envoyer aux clients que la nuit a commencé
		event := events.NewNightEvent()
		s.eventService.SendEventToGame(event, game.ID)

		// TODO: Imlémenter le systeme de tours avec timer pour que les gens jouent la nuit et utilisent leurs capatités
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
	for _, playerID := range game.Players {
		actualPlayer, _ := s.playerRepo.GetPlayerByID(playerID)
		if actualPlayer.Alive() {
			actualPlayerRole := *actualPlayer.Role()
			for _, clan := range actualPlayerRole.GetClans() {
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
		return true, nil
	case werewolves >= villagers:
		// Victoire des loups-garous
		game.Status = entities.GameStatusEnded
		return true, nil
	}

	return false, nil
}
