package slack

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/eparis/bugtool/pkg/config"

	"github.com/ghodss/yaml"
	slackgo "github.com/slack-go/slack"
	"github.com/spf13/cobra"
)

var peopleWithWrongSlackEmail = map[string]string{
	"sttts@redhat.com":       "sschiman@redhat.com",
	"rphillips@redhat.com":   "rphillip@redhat.com",
	"adam.kaplan@redhat.com": "adkaplan@redhat.com",
	"wking@redhat.com":       "trking@redhat.com",
	"sanchezl@redhat.com":    "lusanche@redhat.com",
}

var (
	backOffRegexp = regexp.MustCompile(`slack rate limit exceeded, retry after ([0-9]+s)`)
)

type ChannelClient interface {
	MessageChannel(channel, message string) error
	MessageDebug(message string) error
	MessageEmail(email, message string) error
}

type slackClient struct {
	client       *slackgo.Client
	debugChannel string
	debug        bool
}

func BugzillaToSlackEmail(originalEmail string) string {
	realEmail, ok := peopleWithWrongSlackEmail[originalEmail]
	if ok {
		return realEmail
	}
	return originalEmail
}

func (c *slackClient) MessageDebug(message string) error {
	return c.MessageChannel(c.debugChannel, message)
}

func (c *slackClient) MessageChannel(channel, message string) error {
	if c.debug && channel != c.debugChannel {
		debugMsg := fmt.Sprintf("DEBUG sendto: %s: %s", channel, message)
		return c.MessageDebug(debugMsg)
	}
	_, _, err := c.client.PostMessage(channel, slackgo.MsgOptionText(message, false))
	if err != nil {
		matches := backOffRegexp.FindStringSubmatch(err.Error())
		if len(matches) != 2 {
			return err
		}
		delay, parseErr := time.ParseDuration(matches[1])
		if parseErr != nil {
			return err
		}
		time.Sleep(delay)
		_, _, err = c.client.PostMessage(channel, slackgo.MsgOptionText(message, false))
	}
	return err
}

func (c *slackClient) MessageEmail(email, message string) error {
	if c.debug {
		return c.MessageDebug(fmt.Sprintf("DEBUG: %q will receive:\n%s", email, message))
	}
	user, err := c.client.GetUserByEmail(BugzillaToSlackEmail(email))
	if err != nil {
		return err
	}
	_, _, chanID, err := c.client.OpenIMChannel(user.ID)
	if err != nil {
		return err
	}
	_, _, err = c.client.PostMessage(chanID, slackgo.MsgOptionText(message, false))
	return err
}

type SlackCredentials struct {
	SlackToken             string `yaml:"slackToken"`
	SlackVerificationToken string `yaml:"slackVerificationToken"`
}

func (b SlackCredentials) DecodedSlackToken() string {
	return config.Decode(b.SlackToken)
}

func (b SlackCredentials) DecodedSlackVerificationToken() string {
	return config.Decode(b.SlackVerificationToken)
}

func NewChannelClient(cmd *cobra.Command, ctx context.Context, debugChannel string, debug bool) (ChannelClient, error) {
	b, err := config.GetBytes(cmd, "slack-key", ctx)
	if err != nil {
		return nil, err
	}

	sc := &SlackCredentials{}
	if err := yaml.Unmarshal(b, sc); err != nil {
		return nil, err
	}

	client := slackgo.New(sc.DecodedSlackToken(), slackgo.OptionDebug(true))

	// This slack client is used for production notifications
	// Be careful, this can spam people!
	c := &slackClient{
		client:       client,
		debugChannel: debugChannel,
		debug:        debug,
	}
	return c, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("slack-key", "slackKey", "path containing credentials to use slack")
}
