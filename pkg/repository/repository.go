package repository

import (
	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/tarantool/go-tarantool/v2"
)

type Voting interface {
	CreateVoting(voting votingbot.Voting) (int, error)
	GetAnswers(votingId int) ([]votingbot.Answer, error)
	Vote(ids []int) error
	StopVoting(userId string, votingId int) error
	DeleteVoting(userId string, votingId int) error
}

type Repository struct {
	Voting
	Errors
}

func NewRepository(db *tarantool.Connection) *Repository {
	return &Repository{
		Voting: NewVoting(db),
		Errors: Errors{
			VotingExpired: errVotingExpired,
			WrongUserId:   errWrongUserId},
	}
}
