package votingbot

import (
	"net/url"
)

type MattermostConfig struct {
	MattermostUserName string
	MattermostTeamName string
	MattermostToken    string
	MattermostChannel  string
	MattermostServer   *url.URL
}

type TarantoolConfig struct {
	TarantoolAddress string
	TarantoolUsername string
	TarantoolPassword string
}