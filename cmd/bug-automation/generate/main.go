package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eparis/bugzilla"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/openshift/bugzilla-tools/pkg/api"
)

var (
	testBugs = []string{}
)

func blockerSeverity() []string {
	return []string{
		"unspecified",
		"urgent",
		"high",
		"medium",
	}
}

func onEngineeringStatus() []string {
	return []string{
		"NEW",
		"ASSIGNED",
		"POST",
		"ON_DEV",
	}
}

func allOpenStatus() []string {
	return append(onEngineeringStatus(), "MODIFIED", "ON_QA")
}

func defaultQuery() bugzilla.Query {
	query := bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         onEngineeringStatus(),
		Advanced: []bugzilla.AdvancedQuery{
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Documentation",
				Negate: true,
			},
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Migration Tooling",
				Negate: true,
			},
			{
				Field:  "component",
				Op:     "equals",
				Value:  "odo",
				Negate: true,
			},
		},
		IncludeFields: []string{"id"},
	}
	if len(testBugs) > 0 {
		query.BugIDs = testBugs
		query.BugIDsType = "anyexact"
	}
	return query
}

func bugsTargetOldZeroQuery() bugzilla.Query {
	query := defaultQuery()
	query.TargetRelease = []string{"4.1.0", "4.2.0", "4.3.0", "4.4.0", "4.5.0", "4.6.0"}
	return query
}

func bugsTargetOldZeroUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{
		TargetRelease: "---",
		Comment: &bugzilla.BugComment{
			Private: true,
			Body:    `Unsetting the target release because this bug targets a .0 release which has already shipped. For example it may target 4.2.0. Since 4.2.0 has already shipped such bugs may instead wish to target a future 4.2.z.`,
		},
	}
}

func bugsWithUpcomingSprintQuery() bugzilla.Query {
	query := defaultQuery()
	query.Status = allOpenStatus()
	query.Keywords = []string{"UpcomingSprint"}
	query.KeywordsType = "allwords"
	return query
}

func bugsWithUpcomingSprintUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{
		Keywords: &bugzilla.BugKeywords{
			Remove: []string{"UpcomingSprint"},
		},
		MinorUpdate: true,
	}
}

func bugsWithoutZQuery() bugzilla.Query {
	query := defaultQuery()
	query.Keywords = []string{"Security"}
	query.KeywordsType = "nowords"
	query.Advanced = append(query.Advanced, []bugzilla.AdvancedQuery{
		{
			Field: "dependson",
			Op:    "isempty",
		},
		{
			Field: "target_release",
			Op:    "regexp",
			Value: `^4\.[0-9]+\.z$`,
		},
		{
			Field:  "component",
			Op:     "equals",
			Value:  "Release",
			Negate: true,
		},
	}...)
	return query
}

func bugsWithoutZUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{
		TargetRelease: "---",
		Comment: &bugzilla.BugComment{
			Private: true,
			Body: `This bug sets Target Release equal to a z-stream but has no bug in the 'Depends On' field. As such this is not a valid bug state and the target release is being unset.

Any bug targeting 4.1.z must have a bug targeting 4.2 in 'Depends On.'
Similarly, any bug targeting 4.2.z must have a bug with Target Release of 4.3 in 'Depends On.'`,
		},
	}
}

func targetReleaseWithoutSeverityQuery() bugzilla.Query {
	query := defaultQuery()
	query.Severity = []string{"unspecified"}
	query.Advanced = append(query.Advanced, bugzilla.AdvancedQuery{
		Field:  "target_release",
		Op:     "equals",
		Value:  "---",
		Negate: true,
	})
	return query
}

func targetReleaseWithoutSeverityUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{
		TargetRelease: "---",
		Comment: &bugzilla.BugComment{
			Private: true,
			Body:    `This bug has set a target release without specifying a severity. As part of triage when determining the importance of bugs a severity should be specified. Since these bugs have not been properly triaged we are removing the target release. Teams will need to add a severity before setting the target release again.`,
		},
	}
}

func needsBlockerFlagQuery() bugzilla.Query {
	query := defaultQuery()
	query.Advanced = append(query.Advanced, bugzilla.AdvancedQuery{
		Field:  "flagtypes.name",
		Op:     "substring",
		Value:  "blocker",
		Negate: true,
	})
	return query
}

func bugsNeedBlockerFlagSeverityQuery() bugzilla.Query {
	query := needsBlockerFlagQuery()
	query.Severity = []string{"high", "urgent"}
	return query
}

func bugsNeedBlockerFlagPriorityQuery() bugzilla.Query {
	query := needsBlockerFlagQuery()
	query.Priority = []string{"high", "urgent"}
	return query
}

func bugsNeedBlockerFlagAction() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{
		Flags: []bugzilla.FlagChange{
			{
				Name:   "blocker",
				Status: "?",
			},
		},
		MinorUpdate: true,
	}

}

func blockerPlusWithoutTargetReleaseQuery() bugzilla.Query {
	query := defaultQuery()
	query.Advanced = append(query.Advanced, bugzilla.AdvancedQuery{
		Field: "flagtypes.name",
		Op:    "substring",
		Value: "blocker+",
	})
	query.TargetRelease = []string{"---"}
	return query
}

func blockerPlusWithoutTargetReleaseAction() bugzilla.BugUpdate {
	update := bugsNeedBlockerFlagAction()
	update.MinorUpdate = false
	update.Comment = &bugzilla.BugComment{
		Private: true,
		Body: `This bug sets Target Release equal to a z-stream but has no bug in the 'Depends On' field. As such this is not a valid bug state and the target release is being unset.

Any bug targeting 4.1.z must have a bug targeting 4.2 in 'Depends On.'
Similarly, any bug targeting 4.2.z must have a bug with Target Release of 4.3 in 'Depends On.'`,
	}
	return update
}

func noUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{}
}

type Action api.BugAction

func (a *Action) yaml() string {
	out, err := yaml.Marshal(a)
	if err != nil {
		return ""
	}
	return string(out)
}

func (a Action) write() error {
	actionYaml := a.yaml()
	filename := fmt.Sprintf("../operations/%s.yaml", a.Name)
	return ioutil.WriteFile(filename, []byte(actionYaml), 0644)
}

func doGenerate() error {
	actions := []Action{
		{
			Name:        "targetReleaseWithoutSeverity",
			Description: "Bugs Setting Target Release Without Severity Set",
			Query:       targetReleaseWithoutSeverityQuery(),
			Update:      targetReleaseWithoutSeverityUpdate(),
			Default:     true,
		},
		{
			Name:        "zNoDepends",
			Description: "Z-Stream Bugs With No Depends On",
			Query:       bugsWithoutZQuery(),
			Update:      bugsWithoutZUpdate(),
			Default:     true,
		},
		{
			Name:        "removeUpcomingSprint",
			Description: "Remove UpcomingSprint from all bugs",
			Query:       bugsWithUpcomingSprintQuery(),
			Update:      bugsWithUpcomingSprintUpdate(),
			Default:     false,
		},
		{
			Name:        "bugsTargetOldZero",
			Description: "Open bugs which target closed releases",
			Query:       bugsTargetOldZeroQuery(),
			Update:      bugsTargetOldZeroUpdate(),
			Default:     true,
		},
		{
			Name:        "bugsNeedBlockerFlagSeverity",
			Description: "All bugs that should have at least blocker? based on the severity",
			Query:       bugsNeedBlockerFlagSeverityQuery(),
			Update:      bugsNeedBlockerFlagAction(),
			Default:     true,
		},
		{
			Name:        "bugsNeedBlockerFlagPriority",
			Description: "All bugs that should have at least blocker? based on the priority",
			Query:       bugsNeedBlockerFlagPriorityQuery(),
			Update:      bugsNeedBlockerFlagAction(),
			Default:     true,
		},
		{
			Name:        "blockerPlusWithoutTargetRelease",
			Description: "All bugs that set blocker+ must also set a TargetRelease",
			Query:       blockerPlusWithoutTargetReleaseQuery(),
			Update:      blockerPlusWithoutTargetReleaseAction(),
			Default:     true,
		},
	}

	for _, action := range actions {
		if err := action.write(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := doGenerate()
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Flags().StringSliceVar(&testBugs, "test-bugs", testBugs, "Limit queries to only these specific bugs (CSV)")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
