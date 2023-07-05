package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kelseyhightower/envconfig"
	"github.com/slack-go/slack"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/mlafeldt/listkeeper/functions/internal/data"
)

type output struct {
	Header string
	Text   string
	Footer string
}

type handler struct {
	table         data.TableAPI
	slackUsername string
	slackIconURL  string
}

func main() {
	var env struct {
		TableName     string `envconfig:"TABLE_NAME" required:"true"`
		SlackUsername string `envconfig:"SLACK_USERNAME" required:"true"`
		SlackIconURL  string `envconfig:"SLACK_ICON_URL" required:"true"`
	}
	envconfig.MustProcess("", &env)

	sess := session.Must(session.NewSession())
	h := handler{
		table:         data.NewTable(sess, env.TableName),
		slackUsername: env.SlackUsername,
		slackIconURL:  env.SlackIconURL,
	}

	lambda.Start(h.handle)
}

func (h *handler) handle(ctx context.Context, event *data.FollowerEvent) (*output, error) {
	log.SetPrefix(event.UserID + " ")
	log.Printf("event = %+v", event)

	user, err := h.table.GetUser(ctx, event.UserID)
	if err != nil {
		return nil, err
	}

	log.Printf("slack = %+v", user.Slack)

	var (
		follower = event.Follower
		p        = message.NewPrinter(language.English)
	)

	header := map[string]string{
		"NEW":  "New follower",
		"LOST": "Lost follower",
	}[event.FollowerState]

	text := map[string]string{
		data.FollowerStateReasonFollowed:   p.Sprintf("%s (<https://twitter.com/%s|@%s>) followed you :tada:", follower.Name, follower.Handle, follower.Handle),
		data.FollowerStateReasonUnfollowed: p.Sprintf("%s (<https://twitter.com/%s|@%s>) unfollowed you", follower.Name, follower.Handle, follower.Handle),
		data.FollowerStateReasonDeleted:    p.Sprintf("User with ID %s was deleted", follower.ID),
		data.FollowerStateReasonSuspended:  p.Sprintf("User with ID %s was suspended", follower.ID),
	}[event.FollowerStateReason]

	const sep = "\n\n"
	text += sep
	if follower.Bio != "" {
		text += p.Sprintf("*Bio:* %s%s", follower.Bio, sep)
	}
	if follower.Location != "" {
		text += p.Sprintf("*Location:* %s%s", follower.Location, sep)
	}
	if follower.Name != "" {
		text += p.Sprintf("*Followers:* %d%s", follower.TotalFollowers, sep)
	}

	footer := p.Sprintf("You (@%s) now have %d Twitter followers", user.Handle, event.TotalFollowers)

	if user.Slack.Enabled {
		var accessory *slack.Accessory
		if follower.ProfileImageURL != "" {
			imageURL := strings.Replace(follower.ProfileImageURL, "_normal.", "_400x400.", 1)
			accessory = slack.NewAccessory(slack.NewImageBlockElement(imageURL, "profile image"))
		}

		blocks := []slack.Block{
			slack.NewHeaderBlock(
				slack.NewTextBlockObject("plain_text", header, false, false),
			),
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", text, false, false),
				nil,
				accessory,
			),
			slack.NewContextBlock(
				"",
				slack.NewTextBlockObject("mrkdwn", footer, false, false),
			),
		}

		msg := slack.WebhookMessage{
			Username: h.slackUsername,
			IconURL:  h.slackIconURL,
			Channel:  user.Slack.Channel,
			Blocks:   &slack.Blocks{BlockSet: blocks},
		}

		if err := slack.PostWebhookContext(ctx, user.Slack.WebhookURL, &msg); err != nil {
			return nil, err
		}
	}

	out := output{
		Header: header,
		Text:   text,
		Footer: footer,
	}

	log.Printf("output = %s", out)

	return &out, nil
}
