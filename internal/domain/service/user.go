package service

import (
	"Soloway/internal/domain/entity"
	"fmt"
	"github.com/rs/zerolog"
)

type UsersRepository interface {
	GetAll(spreadsheetID string) ([]entity.User, error)
}

type UserService struct {
	repo   UsersRepository
	logger *zerolog.Logger
}

func NewUserService(repo UsersRepository, logger *zerolog.Logger) *UserService {
	serviceLogger := logger.With().Str("service", "user").Logger()

	return &UserService{
		repo:   repo,
		logger: &serviceLogger,
	}
}

func (s *UserService) GetAll(spreadsheetID string) (users []entity.User, err error) {
	s.logger.Trace().Str("spreadsheetID", spreadsheetID).Msg("GetAll")

	users, err = s.repo.GetAll(spreadsheetID)
	if err != nil {
		return users, fmt.Errorf("UserService.GetAll: %w", err)
	}

	return users, nil
}
