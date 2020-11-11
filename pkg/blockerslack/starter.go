package blockerslack

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"github.com/openshift/bugzilla-tools/pkg/blockerslack/config"
	"github.com/openshift/bugzilla-tools/pkg/blockerslack/reporters/blockers"
	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/slack"
	"github.com/openshift/bugzilla-tools/pkg/teams"
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

	// Be careful, this can spam people!
	slackChannelClient, err := slack.NewChannelClient(cmd, ctx, cfg.SlackDebugChannel, cfg.Debug)
	if err != nil {
		return err
	}

	slackChannelClient.SetEmailMap(cfg.BZToSlackEmail)

	recorder := slack.NewRecorder(slackChannelClient, "BugzillaOperator")

	recorder.Eventf("OperatorStarted", "Bugzilla Operator Started\n\n```\n%s\n```\n", spew.Sdump(cfg))

	schedule := []string{
		"0 12 * * 1-5",
	}
	if cfg.Debug {
		schedule[0] = "* * * * *"
	}
	blockerReporter := blockers.NewBlockersReporter(schedule, cfg, bugData, orgData, slackChannelClient, recorder)

	go blockerReporter.Run(ctx, 1)

	<-ctx.Done()
	return nil
}
