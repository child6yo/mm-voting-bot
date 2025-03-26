package service

import (
	"github.com/child6yo/mm-voting-bot/pkg/app"
)

type Voting interface {
	ListenToEvents()
}

type Service struct {
	Voting
}

func NewService(app app.Application) *Service {
	return &Service{
		Voting: NewVotingServise(app),
	}
}