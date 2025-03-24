package service

import (
	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/mattermost/mattermost-server/v6/model"
)

type Bot interface {
	ListenToEvents()
	sendMsgToTalkingChannel(msg string, replyToId string)
	handleWebSocketEvent(event *model.WebSocketEvent)
	handlePost(post *model.Post)
}

type Service struct {
	Bot
}

func NewService(app app.Application) *Service {
	return &Service{
		Bot: NewBotServise(app),
	}
}