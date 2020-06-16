package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/teams"
	"github.com/spf13/cobra"
)

func doMain(cmd *cobra.Command, _ []string) error {
	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	errs := make(chan error, 1)
	bugData, err := bugs.GetBugData(cmd, orgData, errs)
	if err != nil {
		return err
	}

	// Get All OCP Bugs
	bugs := bugData.GetBugMap()
	targets, err := cmd.Flags().GetStringSlice("target-release")
	if err != nil {
		return err
	}
	bugs = bugs.FilterByTargetRelease(targets)

	severities, err := cmd.Flags().GetStringSlice("severity")
	if err != nil {
		return err
	}
	bugs = bugs.FilterBySeverity(severities)
	keys := make([]string, 0, len(bugs))
	for k := range bugs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, team := range keys {
		teamBugs := bugs[team]
		fmt.Printf("%s,%d\n", team, len(teamBugs))
	}
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)
	cmd.Flags().StringSlice("target-release", []string{"4.5.0", "---"}, "target release to filter by")
	cmd.Flags().StringSlice("severity", []string{"medium", "high", "urgent", "unspecified"}, "severities to filter by")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
