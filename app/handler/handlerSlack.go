package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/StudioAquatan/slack-invite-bot/app/model"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/nlopes/slack"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	// action is used for slack attament action.
	actionAllow = "allow"
	actionDeny  = "deny"
)

type SlackBotInfo struct {
	Client    *slack.Client
	BotID     string
	ChannelID string
}

// interactionHandler handles interactive message response.
type InteractionSlack struct {
	SlackClient       *slack.Client
	VerificationToken string
}

type esaInvitationJson struct {
	Member esaEmailJson `json:"member"`
}

type esaEmailJson struct {
	Emails [] string `json:"emails"`
}

// Send Confirm Message to Slack.
func (s *SlackBotInfo) PostMessageEvent(email string) error {
	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Text:       "入会申請がきたよ！ 承認する？",
		Color:      "#ff4000",
		CallbackID: "invitation",
		Actions: []slack.AttachmentAction{
			{
				Name:  actionAllow + "_" + email,
				Text:  "承認",
				Type:  "button",
				Style: "primary",
			},
			{
				Name:  actionDeny,
				Text:  "拒否",
				Type:  "button",
				Style: "danger",
			},
		},
	}

	// 予め設定しておいたチャンネル宛に送信する
	if _, _, err := s.Client.PostMessage(s.ChannelID, slack.MsgOptionAttachments(attachment)); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}

func (i InteractionSlack) ServeInteractiveSlack(c echo.Context, message *slack.InteractionCallback) (err error) {
	// validation
	// Only accept message from slack with valid token
	if message.Token != i.VerificationToken {
		log.Printf("[ERROR] Invalid token: %s. Verification token is %s", message.Token, i.VerificationToken)
		return c.String(http.StatusUnauthorized, "")
	}

	// Process according to action
	action := message.Actions[0]

	switch strings.Split(action.Name, "_")[0] {
	case actionAllow:
		email := strings.SplitN(action.Name, "_", 2)[1]
		if email == "" {
			log.Printf("[ERROR] var email is empty.")
			title := "emailの取得でエラーが発生しました."
			return responseMessage(c, message.OriginalMessage, title, "")
		}

		err := inviteEsa(email)
		if err != nil {
			log.Printf("[ERROR] Failed to invite to Esa: %s", err)
			title := "Esaの招待作業に失敗しました."
			return responseMessage(c, message.OriginalMessage, title, "")
		}

		err = inviteSlack(email)
		if err != nil {
			log.Printf("[ERROR] Failed to invite to Slack: %s", err)
			title := "Slackの招待作業に失敗しました."
			return responseMessage(c, message.OriginalMessage, title, "")
		}

		title := fmt.Sprintf(":o: @%s さんが入会を承認しました！", message.User.Name)
		return responseMessage(c, message.OriginalMessage, title, "")
	case actionDeny:
		title := fmt.Sprintf(":x: @%s さんが入会を拒否しました．", message.User.Name)
		return responseMessage(c, message.OriginalMessage, title, "")
	default:
		log.Printf("[ERROR] Invalid action was submitted: %s", action.Name)
		return c.String(http.StatusInternalServerError, "")
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(c echo.Context, m slack.Message, title, value string) error {
	m.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	m.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}
	m.ReplaceOriginal = true

	return c.JSON(http.StatusOK, m)
}

func inviteSlack(email string) error {
	var env model.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}
	accessToken := env.SlackAdminToken

	baseUrl := "https://slack.com/api"
	action := "/users.admin.invite"
	endpointUrl := baseUrl + action

	if len(accessToken) <= 0 {
		return errors.New("missing slack_token")
	}

	values := url.Values{}
	values.Set("token", accessToken)
	values.Add("email", email)

	req, err := http.NewRequest(
		"POST",
		endpointUrl,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Printf("POST Slack invitation succeeded! %+v\n", resp)

	return nil
}

func inviteEsa(email string) error {
	var env model.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return err
	}
	accessToken := env.EsaToken

	baseUrl := "https://api.esa.io/v1/"
	action := fmt.Sprintf("teams/%s/invitations", env.EsaTeamName)
	endpointUrl := baseUrl + action + "?access_token=" + accessToken

	if len(accessToken) <= 0 {
		return errors.New("missing esa_token")
	}
	jsonEsa := esaInvitationJson{
		Member: esaEmailJson{
			Emails: []string{email},
		},
	}

	outputJson, err := json.Marshal(&jsonEsa)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		endpointUrl,
		bytes.NewBuffer([]byte(outputJson)),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Printf("POST Esa invitation succeeded! %+v", resp)

	return nil
}
