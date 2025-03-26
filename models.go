package votingbot

type Answer struct {
	Id int
	Description string
	Votes int
}

type Voting struct {
	Id int
	UserId string
	Answers []Answer
}