package repository

import (
	"github.com/child6yo/mm-voting-bot"
	"github.com/tarantool/go-tarantool/v2"
)

type Voting interface {
	CreateVoting(voting votingbot.Voting) (int, error)
	GetVoting(votingId int) ([]votingbot.Answer, error)
}

type Repository struct {
	Voting
}

func NewRepository(db *tarantool.Connection) *Repository {
	return &Repository{
		Voting: NewVoting(db),
	}
}
