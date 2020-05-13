package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if err := bugs.ReconcileBugData(cmd, teams, bugData); err != nil {
		return err
	}
	bugMap := bugData.GetBugMap()

	fmt.Printf("%s,%s,%s,%s,%s,%s\n", "Name", "AllBugs", "UpcomingSprintBugs", "MediumOrHigherSeverity", "4.5Blockers", "Managers")
	for _, team := range teams.Teams {
		name := team.Name
		managers := strings.Join(team.Managers, ",")
		all := bugMap.CountAll(name)
		upcomingSprint := bugMap.CountUpcomingSprint(name)
		medium := bugMap.CountNotLowSeverity(name)
		blockers := bugMap.CountBlocker(name, []string{"---", "4.5.0"})
		fmt.Printf("%s,%d,%d,%d,%d,%s\n", name, all, upcomingSprint, medium, blockers, managers)
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
