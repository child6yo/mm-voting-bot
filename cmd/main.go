package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/child6yo/mm-voting-bot"
	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/child6yo/mm-voting-bot/pkg/service"
	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("env not initialized")
    }

	// Init Application
	config := loadConfig()
	app := app.NewApplication(config)

	app.Logger.Info().Str("config", fmt.Sprint(app.Config)).Msg("")

	setupGracefulShutdown(app)

	// Create a new mattermost client.
	app.MattermostClient = model.NewAPIv4Client(app.Config.MattermostServer.String())

	// Login.
	app.MattermostClient.SetToken(app.Config.MattermostToken)

	if user, resp, err := app.MattermostClient.GetUser("me", ""); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not log in")
	} else {
		app.Logger.Debug().Interface("user", user).Interface("resp", resp).Msg("")
		app.Logger.Info().Msg("Logged in to mattermost")
		app.MattermostUser = user
	}

	// Find and save the bot's team to app struct.
	if team, resp, err := app.MattermostClient.GetTeamByName(app.Config.MattermostTeamName, ""); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not find team. Is this bot a member ?")
	} else {
		app.Logger.Debug().Interface("team", team).Interface("resp", resp).Msg("")
		app.MattermostTeam = team
	}

	// Find and save the talking channel to app struct.
	if channel, resp, err := app.MattermostClient.GetChannelByName(
		app.Config.MattermostChannel, app.MattermostTeam.Id, "",
	); err != nil {
		app.Logger.Fatal().Err(err).Msg("Could not find channel. Is this bot added to that channel ?")
	} else {
		app.Logger.Debug().Interface("channel", channel).Interface("resp", resp).Msg("")
		app.MattermostChannel = channel
	}

	service := service.NewService(*app)
	service.Bot.ListenToEvents()
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

func loadConfig() votingbot.Config {
	var settings votingbot.Config

	settings.MattermostTeamName = os.Getenv("MM_TEAM")
	settings.MattermostUserName = os.Getenv("MM_USERNAME")
	settings.MattermostToken = os.Getenv("MM_TOKEN")
	settings.MattermostChannel = os.Getenv("MM_CHANNEL")
	settings.MattermostServer, _ = url.Parse(os.Getenv("MM_SERVER"))

	return settings
}
