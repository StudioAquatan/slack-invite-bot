package main

import (
	"github.com/StudioAquatan/slack-invite-bot/handler"
	"github.com/StudioAquatan/slack-invite-bot/model"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

type envConfig struct {
	// BotToken is bot user token to access to slack API.
	SlackBotToken string `envconfig:"SLACK_BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	SlackVerificationToken string `envconfig:"SLACK_VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	SlackBotID string `envconfig:"SLACK_BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	SlackChannelID string `envconfig:"SLACK_CHANNEL_ID" required:"true"`

	//Trello Invitation URL
	//TrelloInvitationUrl string `envconfig:"TRELLO_URL" required:"true" default:"trello_url"`
}

func main() {
	e := echo.New()

	// Routing
	// Hello World!
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	//google formからのPOSTを受けてslackに承認を投げる
	e.POST("/slack/", postSlack)

	//slackからの返信を受け取る
	e.POST("/interactive/", interactionSlack)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}

func postSlack(c echo.Context) (err error) {
	// TODO ここわざわざmemberモデル生成しなくてもいい?
	post := new(model.Member)
	if err := c.Bind(post); err != nil {
		return err
	}

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}

	client := slack.New(env.SlackBotToken)
	slackBotInfo := &handler.SlackBotInfo{
		Client:    client,
		BotID:     env.SlackBotID,
		ChannelID: env.SlackChannelID,
	}

	// mailアドレスも一緒におくる
	if err := slackBotInfo.PostMessageEvent(post.Email); err != nil {
		log.Printf("[ERROR] Failed to post message: %s", err)
	}

	return c.String(http.StatusOK, "[INFO] Complete send Message to Slack")
}

func interactionSlack(c echo.Context) (err error) {
	post := new(slack.InteractionCallback)
	if err := c.Bind(post); err != nil { //TODO c.Bind(post)してもpostに値が格納されない
		return err
	}

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}

	client := slack.New(env.SlackBotToken)
	interactionSlack := &handler.InteractionSlack{
		SlackClient:       client,
		VerificationToken: env.SlackVerificationToken,
	}

	return interactionSlack.ServeInteractiveSlack(c, post)
}
