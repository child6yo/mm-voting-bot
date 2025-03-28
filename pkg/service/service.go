package service

import (
	"github.com/child6yo/mm-voting-bot/pkg/app"
)

type Voting interface {
	// ListenToEvents establishes a persistent WebSocket connection to Mattermost,
	// listens for incoming events, and processes them asynchronously. It retries on failure
	// Realizes Voting Bot interface
	ListenToEvents()
}

type Service struct {
	Voting
}

// Returns service instance. Include interfaces: { Voting }
func NewService(app app.Application) *Service {
	return &Service{
		Voting: NewVotingServise(app),
	}
}
