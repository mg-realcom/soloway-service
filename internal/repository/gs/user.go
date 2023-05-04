package gs

import (
	"Soloway/internal/domain/entity"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/api/sheets/v4"
	"strings"
)

const readRange = "// Config!A2:C"

type userRepository struct {
	client sheets.Service
	logger *zerolog.Logger
}

func NewUserRepository(client sheets.Service, logger *zerolog.Logger) *userRepository {
	repoLogger := logger.With().Str("repo", "user").Str("type", "sheets").Logger()

	return &userRepository{
		client: client,
		logger: &repoLogger,
	}
}

func (fr userRepository) GetAll(spreadsheetID string) (users []entity.User, err error) {
	fr.logger.Trace().Msg("GetAll")

	resp, err := fr.client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return users, fmt.Errorf("api error: %w", err)
	}

	for _, row := range resp.Values {
		client, ok := row[0].(string)
		if !ok {
			return users, fmt.Errorf("не могу получить: 'Клиент' в gsheets")
		}

		client = strings.ToLower(client)

		loginStr, ok := row[1].(string)
		if !ok {
			return users, fmt.Errorf("не могу получить: 'Логин' в gsheets")
		}

		login := strings.ToLower(loginStr)

		passStr, ok := row[2].(string)
		if !ok {
			return users, fmt.Errorf("не могу получить: 'Пароль' в gsheets")
		}

		pass := passStr

		user := entity.User{
			Name:     client,
			Login:    login,
			Password: pass,
		}
		users = append(users, user)
	}

	return users, nil
}
