package model

type EnvConfig struct {
	// BotToken is bot user token to access to slack API.
	SlackBotToken string `envconfig:"SLACK_BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	SlackVerificationToken string `envconfig:"SLACK_VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	SlackBotID string `envconfig:"SLACK_BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	SlackChannelID string `envconfig:"SLACK_CHANNEL_ID" required:"true"`

	// When inviting to a workspace Invite as a token-owner's name
	SlackAdminToken string `envconfig:"SLACK_ADMIN_TOKEN" required:"true"`

	EsaToken string `envconfig:"ESA_TOKEN" required:"true"`

	EsaTeamName string `envconfig:"ESA_TEAM_NAME" required:"true"`

	//Trello Invitation URL
	//TrelloInvitationUrl string `envconfig:"TRELLO_URL" required:"true" default:"trello_url"`
}