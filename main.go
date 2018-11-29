package main

import (
	"github.com/StudioAquatan/slack-invite-bot/handler"
	"github.com/StudioAquatan/slack-invite-bot/model"
	"github.com/StudioAquatan/slack-invite-bot/service"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

type envConfig struct {
	// Port is server port to be listened.
	Port string `envconfig:"PORT" default:"3000"`

	// BotToken is bot user token to access to slack API.
	SlackBotToken string `envconfig:"SLACK_BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	SlackVerificationToken string `envconfig:"SLACK_VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	SlackBotID string `envconfig:"SLACK_BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	SlackChannelID string `envconfig:"SLACK_CHANNEL_ID" required:"true"`

	//Esa Wiki Invitation URL
	WikiInvitationUrl string `envconfig:"WIKI_URL" required:"true" default:"wiki_url"`

	//Trello Invitation URL
	TrelloInvitationUrl string `envconfig:"TRELLO_URL" required:"true" default:"trello_url"`
}

func main() {
	// Initialize DB
	InitDB()

	e := echo.New()

	// Routing
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/new/member/", registerMember)
	e.POST("/new/slack/", informMemberSlack)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}

func InitDB() {
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Printf("[ERROR] Invalid action was submitted: %s", err)
		panic("failed to connect database")

	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&model.Member{})
}

func registerMember(c echo.Context) (err error) {
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Printf("[ERROR] Invalid action was submitted: %s", err)
		panic("failed to connect database")
	}
	defer db.Close()

	m := new(model.Member)
	if err = c.Bind(m); err != nil {
		return
	}

	service.CreateMember(db, *m)
	return c.JSON(http.StatusOK, m)
}

func informMemberSlack(c echo.Context) (err error) {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}

	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")
	client := slack.New(env.SlackBotToken)
	slackListener := &handler.SlackListener{
		client:    client,
		botID:     env.SlackBotID,
		channelID: env.SlackChannelID,
	}
	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/slack/interaction", handler.interactionHandler{
		verificationToken: env.SlackVerificationToken,
	})

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	return c.String(http.StatusOK, "[INFO] Complete send Message to Slack") //todo もう少しマシなメッセージを…
}
