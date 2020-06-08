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

	"github.com/eparis/bugtool/pkg/api"
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

func bugsTargetOldZeroQuery() bugzilla.Query {
	return bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         []string{"NEW", "ASSIGNED", "POST", "ON_DEV"},
		TargetRelease:  []string{"4.1.0", "4.2.0", "4.3.0", "4.4.0"},
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
	return bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         []string{"NEW", "ASSIGNED", "POST", "ON_DEV", "MODIFIED", "ON_QA"},
		//Severity:       blockerSeverity(),
		Keywords:     []string{"UpcomingSprint"},
		KeywordsType: "allwords",
		Advanced: []bugzilla.AdvancedQuery{
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Documentation",
				Negate: true,
			},
		},
		IncludeFields: []string{"id"},
	}
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
	return bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         onEngineeringStatus(),
		//Severity:       blockerSeverity(),
		Keywords:     []string{"Security"},
		KeywordsType: "nowords",
		Advanced: []bugzilla.AdvancedQuery{
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
				Value:  "Documentation",
				Negate: true,
			},
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Release",
				Negate: true,
			},
		},
		IncludeFields: []string{"status", "summary", "target_release", "id", "sub_components"},
	}
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
	return bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         onEngineeringStatus(),
		Severity:       []string{"unspecified"},
		Advanced: []bugzilla.AdvancedQuery{
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Documentation",
				Negate: true,
			},
			{
				Field:  "target_release",
				Op:     "equals",
				Value:  "---",
				Negate: true,
			},
		},
		IncludeFields: []string{"id"},
	}
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

func noUpdate() bugzilla.BugUpdate {
	return bugzilla.BugUpdate{}
}

func getAction(name string, desc string, def bool, query bugzilla.Query, update bugzilla.BugUpdate) string {
	bugAction := api.BugAction{
		Default:     def,
		Name:        name,
		Description: desc,
		Query:       query,
		Update:      update,
	}

	out, err := yaml.Marshal(&bugAction)
	if err != nil {
		return ""
	}
	return string(out)
}

func doGenerate() error {
	name := "targetReleaseWithoutSeverity"
	desc := "Bugs Setting Target Release Without Severity Set"
	query := targetReleaseWithoutSeverityQuery()
	update := targetReleaseWithoutSeverityUpdate()
	action := getAction(name, desc, true, query, update)
	filename := fmt.Sprintf("../operations/%s.yaml", name)
	err := ioutil.WriteFile(filename, []byte(action), 0644)
	if err != nil {
		return err
	}

	name = "zNoDepends"
	desc = "Z-Stream Bugs With No Depends On"
	query = bugsWithoutZQuery()
	update = bugsWithoutZUpdate()
	action = getAction(name, desc, true, query, update)
	filename = fmt.Sprintf("../operations/%s.yaml", name)
	err = ioutil.WriteFile(filename, []byte(action), 0644)
	if err != nil {
		return err
	}

	name = "removeUpcomingSprint"
	desc = "Remove UpcomingSprint from all bugs"
	query = bugsWithUpcomingSprintQuery()
	update = bugsWithUpcomingSprintUpdate()
	action = getAction(name, desc, false, query, update)
	filename = fmt.Sprintf("../operations/%s.yaml", name)
	err = ioutil.WriteFile(filename, []byte(action), 0644)
	if err != nil {
		return err
	}

	name = "bugsTargetOldZero"
	desc = "Open bugs which target closed releases"
	query = bugsTargetOldZeroQuery()
	update = bugsTargetOldZeroUpdate()
	action = getAction(name, desc, true, query, update)
	filename = fmt.Sprintf("../operations/%s.yaml", name)
	err = ioutil.WriteFile(filename, []byte(action), 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		RunE: func(_ *cobra.Command, _ []string) error {
			err := doGenerate()
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Execute()
}
