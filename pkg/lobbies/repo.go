package lobbies

import (
	"context"
	"database/sql"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"github.com/google/uuid"
)

type Service struct {
	repo    LobbyRepository
	players auth.PlayerRepository
}

func NewService(r LobbyRepository, p auth.PlayerRepository) *Service {
	return &Service{repo: r, players: p}
}

func (s Service) CreateLobby(ctx context.Context, requesterID string) (string, error) {
	p, errFindPlayer := s.players.FindByID(ctx, requesterID)
	if errFindPlayer != nil {
		return "", errFindPlayer
	}
	players := make([]*auth.Player, 0)
	players = append(players, p)
	l := &Lobby{
		ID:         uuid.NewString(),
		Players:    players,
		CreatedAt:  time.Now(),
		FinishedAt: sql.NullTime{},
	}
	if err := s.repo.Create(ctx, l); err != nil {
		return "", err
	}
	return l.ID, nil
}

func (s Service) StartGame(ctx context.Context, lobbyID string) error {
	return s.repo.StartGame(ctx, lobbyID)
}

func (s Service) StopGame(ctx context.Context, lobbyID string) error {
	t := time.NewTimer(24 * time.Hour)
	<-t.C
	if err := s.repo.FinishGame(ctx, lobbyID); err != nil {
		return err
	}

	return nil
}

func (s Service) JoinGame(ctx context.Context, lobbyID string, playerID string) error {
	p, errFindPlayer := s.players.FindByID(ctx, playerID)
	if errFindPlayer != nil {
		return errFindPlayer
	}

	return s.repo.AddPlayer(ctx, lobbyID, p)
}

func (s Service) FindLobbies(ctx context.Context, playerID string) ([]Lobby, error) {
	return s.repo.Fetch(ctx)
}
