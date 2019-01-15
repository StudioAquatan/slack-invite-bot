# slack-invite-bot
入会後の各サービスへの招待作業の自動化Slack-bot

## Docker image
[![](https://images.microbadger.com/badges/version/studioaquatan/slack-invite-bot.svg)](https://microbadger.com/images/studioaquatan/slack-invite-bot "Get your own version badge on microbadger.com")
[![](https://images.microbadger.com/badges/image/studioaquatan/slack-invite-bot.svg)](https://microbadger.com/images/studioaquatan/slack-invite-bot "Get your own image badge on microbadger.com")

## Development
### 環境変数
実行にあたって以下の環境変数が必要です．

- SLACK_BOT_TOKEN <br>
slack-botのトークン (xoxb-...)

- SLACK_VERIFICATION_TOKEN <br>
slack-botの認証用トークン

- SLACK_BOT_ID <br>
slack-botのID

- SLACK_CHANNEL_ID <br>
投稿先のChannelID

- SLACK_ADMIN_TOKEN <br>
slackの管理者ユーザのlegacyトークン

- ESA_TOKEN <br>
esaの管理者ユーザのトークン

- ESA_TEAM_NAME <br>
esaのチーム名

## Author
- shanpu

##LICENSE
MIT License (c) 2018 StudioAquatan