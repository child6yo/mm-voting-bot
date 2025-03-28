package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/mattermost/mattermost-server/v6/model"
)

type VotingService struct {
	app app.Application
}

func NewVotingServise(app app.Application) *VotingService {
	return &VotingService{app: app}
}

func (b *VotingService) sendMsgToTalkingChannel(msg string, replyToId string) {
	post := &model.Post{}
	post.ChannelId = b.app.MattermostChannel.Id
	post.Message = msg

	post.RootId = replyToId

	if _, _, err := b.app.MattermostClient.CreatePost(post); err != nil {
		b.app.Logger.Error().Err(err).Str("RootID", replyToId).Msg("Failed to create post")
	}
}

func (b *VotingService) ListenToEvents() {
	var err error
	for {
		b.app.MattermostWebSocketClient, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", b.app.Config.MattermostServer.Host+b.app.Config.MattermostServer.Path),
			b.app.MattermostClient.AuthToken,
		)
		if err != nil {
			b.app.Logger.Warn().Err(err).Msg("Mattermost websocket disconnected, retrying")
		}
		b.app.Logger.Info().Msg("Mattermost websocket connected")

		b.app.MattermostWebSocketClient.Listen()

		for event := range b.app.MattermostWebSocketClient.EventChannel {
			go b.handleWebSocketEvent(event)
		}
	}
}

func (b *VotingService) handleWebSocketEvent(event *model.WebSocketEvent) {
	if event.GetBroadcast().ChannelId != b.app.MattermostChannel.Id {
		return
	}

	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		b.app.Logger.Error().Err(err).Msg("Could not cast event to *model.Post")
	}

	if post.UserId == b.app.MattermostUser.Id {
		return
	}

	b.handlePost(post)
}

func (b *VotingService) handlePost(post *model.Post) {
	b.app.Logger.Debug().Str("message", post.Message).Msg("")
	b.app.Logger.Debug().Interface("post", post).Msg("")

	reVoting := regexp.MustCompile(`^!voting`)
	reVote := regexp.MustCompile(`^!vote`)
	reGetVoting := regexp.MustCompile(`^!vshow`)
	reStopVoting := regexp.MustCompile(`^!vstop`)
	reDeleteVoting := regexp.MustCompile(`^!vdelete`)

	var answerAt string
	if post.RootId != "" {
		answerAt = post.RootId
	} else {
		answerAt = post.Id
	}

	switch {
	case reVoting.MatchString(post.Message):
		b.handleVoting(post, answerAt)
		return
	case reVote.MatchString(post.Message):
		b.handleVote(post, answerAt)
		return
	case reGetVoting.MatchString(post.Message):
		b.handleGetVoting(post, answerAt)
		return
	case reStopVoting.MatchString(post.Message):
		b.handleStopVoting(post, answerAt)
		return
	case reDeleteVoting.MatchString(post.Message):
		b.handleDeleteVoting(post, answerAt)
		return
	}
}

func (b *VotingService) handleVoting(post *model.Post, answerAt string) {
	postTokens := parseString(post.Message)
	lenTokens := len(postTokens)
	if lenTokens <= 3 {
		b.sendMsgToTalkingChannel("Use !voting DutationInMinutes Option1 Option2...", post.Id)
		return
	}

	duration, err := strconv.Atoi(postTokens[1])
	if err != nil {
		b.app.Logger.Debug().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Use !voting DutationInMinutes Option1 Option2...", answerAt)
		return
	}

	answers := make([]votingbot.Answer, lenTokens-2)
	for i := 2; i < lenTokens; i++ {
		answer := votingbot.Answer{Id: i - 1, Description: postTokens[i], Votes: 0}
		answers[i-2] = answer
	}

	voting := votingbot.Voting{UserId: post.UserId, Answers: answers, DurationMinutes: duration}

	votingId, err := b.app.Repository.Voting.CreateVoting(voting)
	if err != nil {
		b.app.Logger.Error().Str("error", err.Error()).Msg("")
		return
	}

	for _, a := range answers {
		msg := fmt.Sprintf("Voting ID: %d. Answer ID %d: %s", votingId, a.Id, a.Description)
		b.sendMsgToTalkingChannel(msg, answerAt)
	}
}

func (b *VotingService) handleGetVoting(post *model.Post, answerAt string) {
	postTokens := parseString(post.Message)
	lenTokens := len(postTokens)
	if lenTokens <= 1 || lenTokens > 2 {
		b.sendMsgToTalkingChannel("Use !vshow votingID", answerAt)
		return
	}

	votingId, err := strconv.Atoi(postTokens[1])
	if err != nil {
		b.app.Logger.Debug().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Use !vshow votingID", answerAt)
		return
	}

	answers, err := b.app.Repository.Voting.GetAnswers(votingId)
	if err != nil {
		b.app.Logger.Error().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Seems like voting ID invalid.", answerAt)
		return
	}

	for _, answer := range answers {
		msg := fmt.Sprintf("Voting ID: %d. Answer ID: %d. Answer: %s. Votes: %d",
			votingId, answer.Id, answer.Description, answer.Votes)
		b.sendMsgToTalkingChannel(msg, answerAt)
	}
}

func (b *VotingService) handleVote(post *model.Post, answerAt string) {
	postTokens := parseString(post.Message)
	lenTokens := len(postTokens)
	if lenTokens <= 1 || lenTokens > 3 {
		b.sendMsgToTalkingChannel("Use !vote votingID answerID", answerAt)
		return
	}

	ids := make([]int, 2)
	for i := 1; i <= 2; i++ {
		n, err := strconv.Atoi(postTokens[i])
		if err != nil || n < 0 {
			b.app.Logger.Debug().Str("error", err.Error()).Msg("")
			b.sendMsgToTalkingChannel("Use !vote votingID answerID", answerAt)
			return
		}
		ids[i-1] = n
	}

	err := b.app.Repository.Voting.Vote(ids)
	if err == b.app.Repository.Errors.VotingExpired {
		b.sendMsgToTalkingChannel("Time for this voting has expired.", answerAt)
		return
	} else if err != nil {
		b.app.Logger.Error().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Seems like voting or answer ID invalid.", answerAt)
		return
	}

	b.sendMsgToTalkingChannel("Vote accepted.", answerAt)
}

func (b *VotingService) handleStopVoting(post *model.Post, answerAt string) {
	postTokens := parseString(post.Message)
	lenTokens := len(postTokens)
	if lenTokens <= 1 || lenTokens > 2 {
		b.sendMsgToTalkingChannel("Use !vstop votingID", answerAt)
		return
	}

	votingId, err := strconv.Atoi(postTokens[1])
	if err != nil {
		b.app.Logger.Debug().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Use !vshow votingID", answerAt)
		return
	}

	err = b.app.Repository.Voting.StopVoting(post.UserId, votingId)
	if err == b.app.Repository.Errors.WrongUserId {
		b.sendMsgToTalkingChannel("Only voting owner can stop voting.", answerAt)
		return
	} else if err != nil {
		b.app.Logger.Error().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Seems like votingID invalid.", answerAt)
		return
	}

	b.sendMsgToTalkingChannel("Voting stopped successfully.", answerAt)
}

func (b *VotingService) handleDeleteVoting(post *model.Post, answerAt string) {
	postTokens := parseString(post.Message)
	lenTokens := len(postTokens)
	if lenTokens <= 1 || lenTokens > 2 {
		b.sendMsgToTalkingChannel("Use !vdelete votingID", answerAt)
		return
	}

	votingId, err := strconv.Atoi(postTokens[1])
	if err != nil {
		b.app.Logger.Debug().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Use !vdelete votingID", answerAt)
		return
	}

	err = b.app.Repository.Voting.DeleteVoting(post.UserId, votingId)
	if err == b.app.Repository.Errors.WrongUserId {
		b.sendMsgToTalkingChannel("Only voting owner can delete voting.", answerAt)
		return
	} else if err != nil {
		b.app.Logger.Error().Str("error", err.Error()).Msg("")
		b.sendMsgToTalkingChannel("Seems like votingID invalid.", answerAt)
		return
	}

	b.sendMsgToTalkingChannel("Voting deleted successfully.", answerAt)
}

func parseString(input string) []string {
	re := regexp.MustCompile(`[1-9а-яА-Яa-zA-Z]+(?:\([^()]*\))*`)
	matches := re.FindAllString(input, -1)
	var result []string
	for _, match := range matches {
		match = strings.TrimSpace(match)
		if match != "" {
			result = append(result, match)
		}
	}

	return result
}
