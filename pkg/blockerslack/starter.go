package blockerslack

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	//"github.com/eparis/bugzilla"
	slackgo "github.com/slack-go/slack"
	"github.com/spf13/cobra"

	"github.com/eparis/bugtool/pkg/blockerslack/config"
	"github.com/eparis/bugtool/pkg/blockerslack/reporters/blockers"
	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/slack"
	"github.com/eparis/bugtool/pkg/teams"
)

const bugzillaEndpoint = "https://bugzilla.redhat.com"

func Run(ctx context.Context, cfg config.OperatorConfig, cmd *cobra.Command) error {
	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	bugData, err := bugs.GetBugData(cmd, orgData)
	if err != nil {
		return err
	}

	slackClient := slackgo.New(cfg.Credentials.DecodedSlackToken(), slackgo.OptionDebug(true))

	// This slack client is used for production notifications
	// Be careful, this can spam people!
	slackChannelClient := slack.NewChannelClient(slackClient, cfg.SlackDebugChannel, false)

	recorder := slack.NewRecorder(slackChannelClient, "BugzillaOperator")

	recorder.Eventf("OperatorStarted", "Bugzilla Operator Started\n\n```\n%s\n```\n", spew.Sdump(cfg.Anonymize()))

	schedule := []string{
		//"CRON_TZ=America/New_York 0 7 * * 1-5",
		"* * * * *",
	}
	blockerReporter := blockers.NewBlockersReporter(schedule, cfg, bugData, orgData, slackChannelClient, recorder)

	go blockerReporter.Run(ctx, 1)

	<-ctx.Done()
	return nil
}
