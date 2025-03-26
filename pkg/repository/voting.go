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
	data, err := v.db.Do(tarantool.NewCallRequest("create_voting_with_answers"). // return []interface{}{} ~> int8
	Args([]interface{}{ 
		voting.UserId,
		answers,
	})).Get() 
	if err != nil {
		return 0, err
	}
	
	votingId, ok := data[0].(int8);
	if !ok {
		return 0, fmt.Errorf("votingId encoding error")
	}
		
	return int(votingId), nil
}

func (v *VotingRepo) GetVoting(votingId int) ([]votingbot.Answer, error) {
	voting, err := v.db.Do(tarantool.NewSelectRequest("answers"). // return []interface{}{} ([][]interface{}{} in real)
	Index("voting_idx").										 //	format: [id, votingId, localId, description, votes]
	Iterator(tarantool.IterEq).
	Key([]interface{}{uint(votingId)})).Get()
	if err != nil {
		return nil, err
	}

	var result []votingbot.Answer
	for _, item := range voting {
		tuple, ok := item.([]interface{})
        if !ok || len(tuple) < 4 {
            return nil, fmt.Errorf("invalid tuple format: %v", item)
        }
		result = append(result, votingbot.Answer{
			Id:          int(tuple[2].(int8)),
            Description: tuple[3].(string),
            Votes:       int(tuple[4].(int8)),
		})
	}

	return result, nil
}

func getDescriptions(answers []votingbot.Answer) []string {
    desc := make([]string, len(answers))
    for i, a := range answers {
        desc[i] = a.Description
    }
    return desc
}