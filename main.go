package main

import (
	"encoding/json"
	"github.com/StudioAquatan/slack-invite-bot/handler"
	"github.com/StudioAquatan/slack-invite-bot/model"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

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
	post := c.FormValue("email")
	//if err := c.Bind(post); err != nil {
	//	return err
	//}

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
