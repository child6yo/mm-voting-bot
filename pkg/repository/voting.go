package repository

import (
	"fmt"

	"github.com/child6yo/mm-voting-bot"
	"github.com/tarantool/go-tarantool/v2"
)

type VotingRepo struct {
	db *tarantool.Connection
}

func NewVoting(db *tarantool.Connection) *VotingRepo {
	return &VotingRepo{db: db}
}

func (v *VotingRepo) CreateVoting(voting votingbot.Voting) (int, error) {
	answers := getDescriptions(voting.Answers)
	future := v.db.Do(tarantool.NewCallRequest("create_voting_with_answers").Args([]interface{}{
		voting.UserId,
		answers,
	}))

	data, err := future.Get()
	if err != nil {
		return 0, err
	}
	
	votingId, ok := data[0].(int8);
	if !ok {
		return 0, fmt.Errorf("votingId encoding error")
	}
		
	return int(votingId), nil
}

func (v *VotingRepo) GetVoting() {
	
}

func getDescriptions(answers []votingbot.Answer) []string {
    desc := make([]string, len(answers))
    for i, a := range answers {
        desc[i] = a.Description
    }
    return desc
}