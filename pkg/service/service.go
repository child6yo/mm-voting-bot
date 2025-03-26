package service

import (
	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/mattermost/mattermost-server/v6/model"
)

type Voting interface {
	ListenToEvents()
	sendMsgToTalkingChannel(msg string, replyToId string)
	handleWebSocketEvent(event *model.WebSocketEvent)
	handlePost(post *model.Post)
	handleGetVoting(post *model.Post)
}

type Service struct {
	Voting
}

func NewService(app app.Application) *Service {
	return &Service{
		Voting: NewVotingServise(app),
	}
}