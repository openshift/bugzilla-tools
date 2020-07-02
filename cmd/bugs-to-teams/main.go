package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/teams"
	"github.com/spf13/cobra"
)

func doMain(cmd *cobra.Command, _ []string) error {
	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	bugData, err := bugs.GetBugData(cmd, orgData)
	if err != nil {
		return err
	}

	bugs := bugData.GetTeamMap()
	b, err := json.Marshal(bugs)
	if err != nil {
		return err
	}
	os.Stdout.Write(b)
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
