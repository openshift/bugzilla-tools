package config

import (
	"context"

	"github.com/eparis/bugtool/pkg/config"

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
	return c, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "config.yaml", "Path to operator config")
}
