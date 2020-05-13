package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/eparis/react-material/pkg/bugs"
	"github.com/eparis/react-material/pkg/teams"
	"github.com/spf13/cobra"
)

func doMain(cmd *cobra.Command, _ []string) error {
	teams, err := teams.GetTeamData(cmd)
	if err != nil {
		return err
	}

	bugData := &bugs.BugData{}
	err = bugs.ReconcileBugData(cmd, teams, bugData)
	if err != nil {
		return err
	}

	// Get All OCP Bugs
	bugs := bugData.GetBugMap()
	// Filter by upcoming
	bugs = bugs.FilterByTargetRelease([]string{"4.5.0", "---"})
	bugs = bugs.FilterBySeverity([]string{"medium", "high", "urgent", "unspecified"})
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
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Execute()
}
