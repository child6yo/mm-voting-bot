package votingbot

type Answer struct {
	GlobalId    int
	Id          int
	Description string
	Votes       int
}

type Voting struct {
	Id        int
	UserId    string
	Answers   []Answer
	DurationMinutes int
}