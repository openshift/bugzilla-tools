package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

var peopleWithWrongSlackEmail = map[string]string{
	"sttts@redhat.com":       "sschiman@redhat.com",
	"rphillips@redhat.com":   "rphillip@redhat.com",
	"adam.kaplan@redhat.com": "adkaplan@redhat.com",
	"wking@redhat.com":       "trking@redhat.com",
	"sanchezl@redhat.com":    "lusanche@redhat.com",
}

type ChannelClient interface {
	MessageChannel(channel, message string) error
	MessageDebug(message string) error
	MessageEmail(email, message string) error
}

type slackClient struct {
	client       *slack.Client
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
	_, _, err := c.client.PostMessage(channel, slack.MsgOptionText(message, false))
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
	_, _, err = c.client.PostMessage(chanID, slack.MsgOptionText(message, false))
	return err
}

func NewChannelClient(client *slack.Client, debugChannel string, debug bool) ChannelClient {
	c := &slackClient{
		client:       client,
		debugChannel: debugChannel,
		debug:        debug,
	}
	return c
}
