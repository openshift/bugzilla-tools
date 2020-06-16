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
	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	errs := make(chan error, 1)
	bugData, err := bugs.GetBugData(cmd, orgData, errs)
	if err != nil {
		return err
	}
	bugMap := bugData.GetBugMap()

	fmt.Printf("%s,%s,%s,%s\n", "Name", "AllBugs", "UpcomingSprintBugs", "Managers")
	for _, team := range orgData.Teams {
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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
