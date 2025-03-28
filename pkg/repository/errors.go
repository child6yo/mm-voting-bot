package repository

import "errors"

var (
	errVotingExpired error = errors.New("time for voting has expired")
	errWrongUserId error = errors.New("wrong user id")
)

type Errors struct {
	VotingExpired error
	WrongUserId error
}