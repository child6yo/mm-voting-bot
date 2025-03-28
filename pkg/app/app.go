package app

import (
	"os"
	"time"

	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/child6yo/mm-voting-bot/pkg/repository"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
)

type Application struct {
	Config                    votingbot.MattermostConfig
	Logger                    zerolog.Logger
	Repository                *repository.Repository
	MattermostClient          *model.Client4
	MattermostWebSocketClient *model.WebSocketClient
	MattermostUser            *model.User
	MattermostChannel         *model.Channel
	MattermostTeam            *model.Team
}

func NewApplication(config votingbot.MattermostConfig, repository *repository.Repository) *Application {
	return &Application{
		Config: config,
		Logger: zerolog.New(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC822,
			}).With().Timestamp().Logger(),
		Repository: repository,
	}
}
