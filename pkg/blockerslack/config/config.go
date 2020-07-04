package config

import (
	"context"

	"github.com/eparis/bugtool/pkg/config"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

type Transition struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type OperatorConfig struct {
	Debug             bool              `yaml:"debug"`
	SlackDebugChannel string            `yaml:"slackDebugChannel"`
	BZToSlackEmail    map[string]string `yaml:"bz_to_slack_email"`
}

func GetConfig(cmd *cobra.Command, ctx context.Context) (*OperatorConfig, error) {
	configBytes, err := config.GetBytes(cmd, "config", ctx)
	if err != nil {
		return nil, err
	}
	c := &OperatorConfig{}
	if err := yaml.Unmarshal(configBytes, c); err != nil {
		return nil, err
	}
	return c, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "config.yaml", "Path to operator config")
}
