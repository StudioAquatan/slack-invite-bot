package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/StudioAquatan/slack-invite-bot/model"
	"github.com/jinzhu/gorm"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	// action is used for slack attament action.
	actionAllow = "allow"
	actionDeny  = "deny"
)

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
}

type esaInvitationJson struct {
	Member esaEmailJson `json:"member"`
}

type esaEmailJson struct {
	Emails [] string `json:"emails"`
}

type slackInvitationJson struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// ListenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Text:       "入会申請がきたよ！ 承認する？",
		Color:      "#ff4000",
		CallbackID: "invitation",
		Actions: []slack.AttachmentAction{
			{
				Name:  actionAllow,
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

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}

	if _, _, err := s.client.PostMessage(ev.Channel, "", params); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// validation
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Process according to action
	action := message.Actions[0]
	switch action.Name {
	case actionAllow:
		// todo 承認できる人間を限定する？
		db, err := gorm.Open("sqlite3", "../db.sqlite3")
		if err != nil {
			log.Printf("[ERROR] Invalid action was submitted: %s", err)
			panic("failed to connect database")
		}
		defer db.Close()

		var member model.Member
		db.Where("process = ?", "0").Last(&member)

		// todo invite関数を並列処理する
		err = inviteEsa(member.Email)
		if err != nil {
			log.Printf("[ERROR] Failed to invite to Esa: %s", err)
			title := "Esaの招待作業に失敗しました."
			responseMessage(w, message.OriginalMessage, title, "")
			return
		}
		db.Model(&member).Update("process", member.Process+1) //todo serviceにまとめる

		err = inviteSlack(member.Email)
		if err != nil {
			log.Printf("[ERROR] Failed to invite to Slack: %s", err)
			title := "Slackの招待作業に失敗しました."
			responseMessage(w, message.OriginalMessage, title, "")
			return
		}
		db.Model(&member).Update("process", member.Process+1) //todo serviceにまとめる

		title := fmt.Sprintf(":o: @%s さんが入会を承認しました！", message.User.Name)
		responseMessage(w, message.OriginalMessage, title, "")
		return
	case actionDeny:
		// todo 拒否できる人間を限定する？
		// todo 断ったときはどうするか 一度保留にしておくorデータベースに情報を残したまま放置orデータも消す．
		title := fmt.Sprintf(":x: @%s さんが入会を拒否しました．", message.User.Name)
		responseMessage(w, message.OriginalMessage, title, "")
		return
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

func inviteSlack(email string) error {
	baseUrl := "https://slack.com/api"
	action := "/users.admin.invite"
	accessToken := os.Getenv("SLACK_TOKEN") //todo envconfigにまとめる

	endpointUrl := baseUrl + action

	if len(accessToken) > 0 {
		jsonSlack := slackInvitationJson{
			Token: accessToken,
			Email: email,
		}

		outputJson, err := json.Marshal(&jsonSlack)
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
		log.Printf("POST Slack invitation succeeded! %s", fmt.Sprintf("%s", resp)) //TODO ここのstringへの変換
	} else {
		log.Printf("[ERROR] Can't find \"Slack_TOKEN\".")
	}

	return nil
}

func inviteEsa(email string) error {
	baseUrl := "https://api.esa.io/v1/"
	action := fmt.Sprintf("teams/%s/invitations", os.Getenv("ESA_TEAMNAME")) //todo envconfigにまとめる
	accessToken := os.Getenv("ESA_TOKEN")

	endpointUrl := baseUrl + action + "?" + accessToken

	if len(accessToken) > 0 {
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
		log.Printf("POST Esa invitation succeeded! %s", fmt.Sprintf("%s", resp)) //TODO ここのstringへの変換
	} else {
		log.Printf("[ERROR] Can't find \"ESA_TOKEN\".")
	}

	return nil
}
