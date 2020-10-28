package config

import (
	"context"

	"github.com/openshift/bugzilla-tools/pkg/config"

	"github.com/spf13/cobra"
)

type Transition struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type OperatorConfig struct {
	Debug             bool              `json:"debug"`
	SlackDebugChannel string            `json:"slackDebugChannel"`
	BZToSlackEmail    map[string]string `json:"bz_to_slack_email"`
}

func GetConfig(cmd *cobra.Command, ctx context.Context) (*OperatorConfig, error) {
	c := &OperatorConfig{}
	err := config.GetConfig(cmd, "config", ctx, c)
	if err != nil {
		return nil, err
	}
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	c.Debug = c.Debug || debug
	return c, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "config.yaml", "Path to operator config")
}
