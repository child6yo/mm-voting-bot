package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"

	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/child6yo/mm-voting-bot/pkg/repository"
	"github.com/child6yo/mm-voting-bot/pkg/service"
	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	// Init Application
	mmConfig, ttConfig := loadConfig()
	conn, err := repository.CreateTarantoolDb(ttConfig)
	if err != nil {
		fmt.Println(err)
	}
	addr := conn.Addr()

	repository := repository.NewRepository(conn)
	app := app.NewApplication(mmConfig, repository)
	app.Logger.Info().Str("config", fmt.Sprint(app.Config)).Msg("")
	app.Logger.Info().Str("tarantool address", addr.String()).Msg("")
	setupGracefulShutdown(app)

	// Create a new mattermost client
	app.MattermostClient = model.NewAPIv4Client(app.Config.MattermostServer.String())

	// Login
	app.MattermostClient.SetToken(app.Config.MattermostToken)

	if user, resp, err := app.MattermostClient.GetUser("me", ""); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not log in")
	} else {
		app.Logger.Debug().Interface("user", user).Interface("resp", resp).Msg("")
		app.Logger.Info().Msg("Logged in to mattermost")
		app.MattermostUser = user
	}

	// Find and save the bot's team to app struct
	if team, resp, err := app.MattermostClient.GetTeamByName(app.Config.MattermostTeamName, ""); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not find team. Is this bot a member ?")
	} else {
		app.Logger.Debug().Interface("team", team).Interface("resp", resp).Msg("")
		app.MattermostTeam = team
	}

	// Find and save the talking channel to app struct
	if channel, resp, err := app.MattermostClient.GetChannelByName(
		app.Config.MattermostChannel, app.MattermostTeam.Id, "",
	); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not find channel. Is this bot added to that channel ?")
	} else {
		app.Logger.Debug().Interface("channel", channel).Interface("resp", resp).Msg("")
		app.MattermostChannel = channel
	}

	// Bot start
	service := service.NewService(*app)
	service.Voting.ListenToEvents()
}

func setupGracefulShutdown(app *app.Application) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if app.MattermostWebSocketClient != nil {
				app.Logger.Info().Msg("Closing websocket connection")
				app.MattermostWebSocketClient.Close()
			}
			app.Logger.Info().Msg("Shutting down")
			os.Exit(0)
		}
	}()
}

func loadConfig() (votingbot.MattermostConfig, votingbot.TarantoolConfig) {
	var mmSettings votingbot.MattermostConfig

	mmSettings.MattermostTeamName = os.Getenv("BOT_MM_TEAM")
	mmSettings.MattermostUserName = os.Getenv("BOT_MM_USERNAME")
	mmSettings.MattermostToken = os.Getenv("BOT_MM_TOKEN")
	mmSettings.MattermostChannel = os.Getenv("BOT_MM_CHANNEL")
	mmSettings.MattermostServer, _ = url.Parse(os.Getenv("BOT_MM_SERVER"))

	var ttSettings votingbot.TarantoolConfig

	ttSettings.TarantoolAddress = os.Getenv("BOT_TT_ADDRES")
	ttSettings.TarantoolUsername = os.Getenv("BOT_TT_USERNAME")
	ttSettings.TarantoolPassword = os.Getenv("BOT_TT_PASSWORD")

	return mmSettings, ttSettings
}
