package entities_test

import (
	"shamus-backend/internal/domain/entities"
	"testing"
)

func TestGame_CanStart(t *testing.T) {
	tests := []struct {
		name    string
		game    *entities.Game
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid game configuration",
			game: &entities.Game{
				ID:     "test-game",
				Status: entities.GameStatusWaiting,
				Phase:  entities.PhaseStart,
				Players: []entities.PlayerID{
					"player1", "player2", "player3", "player4",
					"player5", "player6", "player7", "player8",
				},
				Settings: entities.GameSettings{
					Roles: map[entities.RoleType]int{
						entities.RoleVillager: 4,
						entities.RoleWerewolf: 2,
						entities.RoleSeer:     1,
						entities.RoleWitch:    1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "not enough players",
			game: &entities.Game{
				ID:      "test-game",
				Players: []entities.PlayerID{"player1", "player2"},
				Settings: entities.GameSettings{
					Roles: map[entities.RoleType]int{
						entities.RoleVillager: 1,
						entities.RoleWerewolf: 1,
					},
				},
			},
			wantErr: true,
			errMsg:  "not enough players",
		},
		{
			name: "role count mismatch",
			game: &entities.Game{
				ID: "test-game",
				Players: []entities.PlayerID{
					"player1", "player2", "player3", "player4",
				},
				Settings: entities.GameSettings{
					Roles: map[entities.RoleType]int{
						entities.RoleVillager: 2,
						entities.RoleWerewolf: 1,
					},
				},
			},
			wantErr: true,
			errMsg:  "role count mismatch",
		},
		{
			name: "too many seers",
			game: &entities.Game{
				ID: "test-game",
				Players: []entities.PlayerID{
					"player1", "player2", "player3", "player4",
				},
				Settings: entities.GameSettings{
					Roles: map[entities.RoleType]int{
						entities.RoleVillager: 1,
						entities.RoleWerewolf: 1,
						entities.RoleSeer:     2, // Max 1 allowed
					},
				},
			},
			wantErr: true,
			errMsg:  "too many",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.game.CanStart()
			if (err != nil) != tt.wantErr {
				t.Errorf("Game.CanStart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Game.CanStart() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGameSettings_TotalRoles(t *testing.T) {
	settings := entities.GameSettings{
		Roles: map[entities.RoleType]int{
			entities.RoleVillager: 4,
			entities.RoleWerewolf: 2,
			entities.RoleSeer:     1,
			entities.RoleWitch:    1,
		},
	}

	total := settings.TotalRoles()
	expected := 8

	if total != expected {
		t.Errorf("GameSettings.TotalRoles() = %d, want %d", total, expected)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
