package service

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/child6yo/mm-voting-bot/pkg/app"
	"github.com/mattermost/mattermost-server/v6/model"
)

type BotService struct {
	app app.Application
}

func NewBotServise(app app.Application) *BotService {
	return &BotService{app: app}
}


func (b *BotService) sendMsgToTalkingChannel(msg string, replyToId string) {
	// Note that replyToId should be empty for a new post.
	// All replies in a thread should reply to root.

	post := &model.Post{}
	post.ChannelId = b.app.MattermostChannel.Id
	post.Message = msg

	post.RootId = replyToId

	if _, _, err := b.app.MattermostClient.CreatePost(post); err != nil {
		b.app.Logger.Error().Err(err).Str("RootID", replyToId).Msg("Failed to create post")
	}
}

func (b *BotService) ListenToEvents() {
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
			// Launch new goroutine for handling the actual event.
			// If required, you can limit the number of events beng processed at a time.
			go b.handleWebSocketEvent(event)
		}
	}
}

func (b *BotService) handleWebSocketEvent(event *model.WebSocketEvent) {

	// Ignore other channels.
	if event.GetBroadcast().ChannelId != b.app.MattermostChannel.Id {
		return
	}

	// Ignore other types of events.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// Since this event is a post, unmarshal it to (*model.Post)
	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		b.app.Logger.Error().Err(err).Msg("Could not cast event to *model.Post")
	}

	// Ignore messages sent by this bot itself.
	if post.UserId == b.app.MattermostUser.Id {
		return
	}

	// Handle however you want.
	b.handlePost(post)
}

func (b *BotService) handlePost(post *model.Post) {
	b.app.Logger.Debug().Str("message", post.Message).Msg("")
	b.app.Logger.Debug().Interface("post", post).Msg("")

	if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {

		// If post has a root ID then its part of thread, so reply there.
		// If not, then post is independent, so reply to the post.
		if post.RootId != "" {
			b.sendMsgToTalkingChannel("I replied in an existing thread.", post.RootId)
		} else {
			b.sendMsgToTalkingChannel("I just replied to a new post, starting a chain.", post.Id)
		}
		return
	}
}