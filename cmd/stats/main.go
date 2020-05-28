package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eparis/bugzilla"
	"github.com/eparis/react-material/pkg/bugs"
	"github.com/eparis/react-material/pkg/teams"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

func countUpcomingMediumAndHigher(bugs []*bugzilla.Bug) int {
	severities := sets.NewString("medium", "high", "urgent", "unspecified")
	count := 0
	for _, bug := range bugs {
		keywords := sets.NewString(bug.Keywords...)
		if !keywords.Has("UpcomingSprint") {
			continue
		}
		if !severities.Has(bug.Severity) {
			continue
		}
		count += 1
	}
	return count
}

func countUpcomingBlocker(bugs []*bugzilla.Bug) int {
	severities := sets.NewString("medium", "high", "urgent", "unspecified")
	targetReleases := sets.NewString("4.5.0", "---")
	count := 0
	for _, bug := range bugs {
		keywords := sets.NewString(bug.Keywords...)
		if !keywords.Has("UpcomingSprint") {
			continue
		}
		if !severities.Has(bug.Severity) {
			continue
		}
		if !targetReleases.Has(bug.TargetRelease[0]) {
			continue
		}
		count += 1

	}
	return count
}

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

	fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n", "Name", "AllBugs", "UpcomingSprintBugs", "MediumOrHigherSeverity", "UpcomingMedium", "4.5Blockers", "UpcomingBlocker", "Managers")
	for _, team := range teams.Teams {
		name := team.Name
		managers := strings.Join(team.Managers, ",")
		bugs := bugMap[name]
		all := bugMap.CountAll(name)
		upcomingSprint := bugMap.CountUpcomingSprint(name)
		medium := bugMap.CountNotLowSeverity(name)
		upcomingMedium := countUpcomingMediumAndHigher(bugs)
		blockers := bugMap.CountBlocker(name, []string{"---", "4.5.0"})
		upcomingBlockers := countUpcomingBlocker(bugs)
		fmt.Printf("%s,%d,%d,%d,%d,%d,%d,%s\n", name, all, upcomingSprint, medium, upcomingMedium, blockers, upcomingBlockers, managers)
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
