package repository

import (
	"fmt"
	"time"

	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
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
			uint(voting.DurationMinutes),
		})).Get()
	if err != nil {
		return 0, err
	}

	votingId, ok := data[0].(int8)
	if !ok {
		return 0, fmt.Errorf("votingId encoding error")
	}

	return int(votingId), nil
}

func (v *VotingRepo) GetAnswers(votingId int) ([]votingbot.Answer, error) {
	answers, err := v.db.Do(tarantool.NewSelectRequest("answers"). // return []interface{}{} ([][]interface{}{} in real)
									Index("voting_idx"). //	format: [id, votingId, localId, description, votes]
									Iterator(tarantool.IterEq).
									Key([]interface{}{uint(votingId)})).Get()
	if err != nil {
		return nil, err
	}
	if len(answers) < 1 {
		return nil, fmt.Errorf("empty data: %v", answers)
	}

	var result []votingbot.Answer
	for _, item := range answers {
		tuple, ok := item.([]interface{})
		if !ok || len(tuple) < 4 {
			return nil, fmt.Errorf("invalid tuple format: %v", item)
		}
		result = append(result, votingbot.Answer{
			GlobalId:    int(tuple[0].(int8)),
			Id:          int(tuple[2].(int8)),
			Description: tuple[3].(string),
			Votes:       int(tuple[4].(int8)),
		})
	}

	return result, nil
}

func (v *VotingRepo) getVoting(votingId int) ([]interface{}, error) {
	voting, err := v.db.Do(tarantool.NewSelectRequest("votings"). // return []interface{}{}
									Iterator(tarantool.IterEq). // format: [id, userId, expiresAt]
									Key([]interface{}{uint(votingId)})).Get()

	if err != nil {
		return nil, err
	}

	if len(voting) < 1 {
		return nil, fmt.Errorf("empty data: %v", voting)
	}

	tuple, ok := voting[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tuple format: %v", voting)
	}

	return tuple, nil
}

func (v *VotingRepo) Vote(ids []int) error {
	voting, err := v.getVoting(ids[0]) // format: [id, userId, expiresAt]

	if err != nil {
		return err
	}

	expiresAt, ok := voting[2].(datetime.Datetime)
	if !ok {
		return fmt.Errorf("error with parsing datetime")
	}

	now := time.Now().UTC()
	fmt.Println(now)
	fmt.Println(expiresAt.ToTime())

	if now.After(expiresAt.ToTime()) {
		return errVotingExpired
	}

	answers, err := v.GetAnswers(ids[0])
	if err != nil {
		return err
	}

	if len(answers) < ids[1]-1 {
		return fmt.Errorf("invalid answer id")
	}

	target := answers[ids[1]-1]
	target.Votes++

	_, err = v.db.Do(
		tarantool.NewUpdateRequest("answers").
			Index("primary").
			Key(tarantool.IntKey{I: target.GlobalId}).
			Operations(tarantool.NewOperations().Assign(4, target.Votes)),
	).Get()
	if err != nil {
		return err
	}

	return nil
}

func (v *VotingRepo) StopVoting(userId string, votingId int) error {
	voting, err := v.getVoting(votingId) // format: [id, userId, expiresAt]

	if err != nil {
		return err
	}

	uId, ok := voting[1].(string)
	if !ok {
		return fmt.Errorf("error with parsing userId")
	}

	if userId != uId {
		return errWrongUserId
	}

	now, err := datetime.MakeDatetime(time.Now().UTC())
	if err != nil {
		return err
	}

	_, err = v.db.Do(
		tarantool.NewUpdateRequest("votings").
			Index("primary").
			Key(tarantool.IntKey{I: votingId}).
			Operations(tarantool.NewOperations().Assign(2, now)),
	).Get()
	if err != nil {
		return err
	}

	return nil
}

func (v *VotingRepo) DeleteVoting(userId string, votingId int) error {
	voting, err := v.getVoting(votingId) // format: [id, userId, expiresAt]

	if err != nil {
		return err
	}

	uId, ok := voting[1].(string)
	if !ok {
		return fmt.Errorf("error with parsing userId")
	}

	if userId != uId {
		return errWrongUserId
	}

	_, err = v.db.Do(tarantool.NewCallRequest("delete_voting_with_answers").
		Args([]interface{}{votingId})).Get()
	if err != nil {
		return err
	}

	return nil
}

func getDescriptions(answers []votingbot.Answer) []string {
	desc := make([]string, len(answers))
	for i, a := range answers {
		desc[i] = a.Description
	}
	return desc
}
