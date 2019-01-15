package main

import (
	"encoding/json"
	"github.com/StudioAquatan/slack-invite-bot/app/handler"
	"github.com/StudioAquatan/slack-invite-bot/app/model"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/nlopes/slack"
	"log"
	"net/http"
)

func main() {
	e := echo.New()

	// Routing
	// Hello World!
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	//POSTを受けるとslackに入会の確認を投げる
	e.POST("/slack/", postSlack)

	//slackからの返信を受け取る
	e.POST("/interactive/", interactionSlack)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}

func postSlack(c echo.Context) (err error) {
	post := c.FormValue("email")

	var env model.EnvConfig
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
	if err := slackBotInfo.PostMessageEvent(post); err != nil {
		log.Printf("[ERROR] Failed to post message: %s", err)
	}

	return c.String(http.StatusOK, "[INFO] Complete send Message to Slack")
}

func interactionSlack(c echo.Context) (err error) {
	post := c.FormValue("payload")

	var data slack.InteractionCallback
	if err := json.Unmarshal([]byte(post), &data); err != nil {
		log.Printf("[ERROR] Failed to process json unmarshal: %s", err)
		return err
	}

	var env model.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}

	client := slack.New(env.SlackBotToken)
	interactionSlack := &handler.InteractionSlack{
		SlackClient:       client,
		VerificationToken: env.SlackVerificationToken,
	}

	return interactionSlack.ServeInteractiveSlack(c, &data)
}
