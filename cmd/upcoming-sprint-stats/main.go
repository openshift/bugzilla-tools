package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	bugMap := bugData.GetTeamMap()

	fmt.Printf("%s,%s,%s,%s\n", "Name", "AllBugs", "UpcomingSprintBugs", "Managers")

	teams := orgData.GetTeamNames()
	for _, team := range teams {
		teamInfo := orgData.Teams[team]
		managers := strings.Join(teamInfo.Managers, ",")
		all := bugMap.CountAll(team)
		upcomingSprint := bugMap.CountUpcomingSprint(team)
		fmt.Printf("%s,%d,%d,%s\n", team, all, upcomingSprint, managers)
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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
