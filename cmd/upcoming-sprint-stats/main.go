package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/teams"
	"github.com/spf13/cobra"
)

func doMain(cmd *cobra.Command, _ []string) error {
	teams, err := teams.GetTeamData(cmd)
	if err != nil {
		return err
	}

	bugData := &bugs.BugData{}
	if err := bugs.ReconcileBugData(cmd, teams, bugData); err != nil {
		return err
	}
	bugMap := bugData.GetBugMap()

	fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n", "Name", "AllBugs", "UpcomingSprintBugs", "Managers")
	for _, team := range teams.Teams {
		name := team.Name
		managers := strings.Join(team.Managers, ",")
		all := bugMap.CountAll(name)
		upcomingSprint := bugMap.CountUpcomingSprint(name)
		fmt.Printf("%s,%d,%d,%s\n", name, all, upcomingSprint, managers)
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
