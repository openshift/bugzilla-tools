package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/teams"
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

	targets, err := cmd.Flags().GetStringSlice("target-release")
	if err != nil {
		return err
	}
	bugData = bugData.FilterByTargetRelease(targets)

	severities, err := cmd.Flags().GetStringSlice("severity")
	if err != nil {
		return err
	}
	bugData = bugData.FilterBySeverity(severities)

	bugs := bugData.GetTeamMap()
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
	cmd.Flags().StringSlice("target-release", bugs.CurrentReleaseTargets, "target release to filter by")
	cmd.Flags().StringSlice("severity", []string{"medium", "high", "urgent", "unspecified"}, "severities to filter by")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
