package bugs

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eparis/bugtool/pkg/teams"
	"github.com/eparis/bugzilla"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	UpcomingSprint   = "UpcomingSprint"
	APIKeyFlagName   = "bugzilla-key"
	apiKeyFlagDefVal = "bugzillaKey"
	apiKeyFlagUsage  = "Path to file containing BZ API key"

	bugDataFlagName   = "test-bug-data"
	bugDataFlagDefVal = ""
	bugDataFlagUsage  = "Path to file containing test bug data"

	CurrentRelease = "4.5.0"
)

type PeopleMap map[string][]*bugzilla.Bug

type TeamMap map[string][]*bugzilla.Bug

func (b TeamMap) Teams() []string {
	out := []string{}
	for team := range b {
		out = append(out, team)
	}
	sort.Strings(out)
	return out
}

func (b TeamMap) CountAll(team string) int {
	return len(b[team])
}

func (b TeamMap) CountUpcomingSprint(team string) int {
	count := 0
	for _, bug := range b[team] {
		for _, found := range bug.Keywords {
			if found == UpcomingSprint {
				count += 1
				break
			}
		}
	}
	return count
}

func (b TeamMap) CountNotUpcomingSprint(team string) int {
	return b.CountAll(team) - b.CountUpcomingSprint(team)
}

func (b TeamMap) CountLowSeverity(team string) int {
	count := 0
	for _, bug := range b[team] {
		if bug.Severity == "low" {
			count += 1
		}
	}
	return count
}

func (b TeamMap) CountNotLowSeverity(team string) int {
	return b.CountAll(team) - b.CountLowSeverity(team)
}

func (b TeamMap) CountTargetRelease(team string, targets []string) int {
	count := 0
	for _, bug := range b[team] {
		targetRelease := bug.TargetRelease
		for _, target := range targets {
			if targetRelease[0] == target {
				count += 1
				break
			}
		}
	}
	return count
}

func (b TeamMap) CountBlocker(team string, targets []string) int {
	count := 0
	for _, bug := range b[team] {
		targetRelease := bug.TargetRelease
		severity := bug.Severity

		if severity == "low" {
			continue
		}
		for _, target := range targets {
			if targetRelease[0] == target {
				count += 1
				break
			}
		}
	}
	return count
}

type BugData struct {
	sync.RWMutex
	bugs    []*bugzilla.Bug
	cmd     *cobra.Command
	client  bugzilla.Client
	query   bugzilla.Query
	orgData *teams.OrgData
}

func (bd *BugData) clone() *BugData {
	bugs := bd.GetBugs()
	newBugs := make([]*bugzilla.Bug, len(bugs))
	copy(newBugs, bugs)

	bugData := &BugData{
		cmd:     bd.cmd,
		client:  bd.client,
		query:   bd.query,
		orgData: bd.orgData,
	}
	bugData.set(newBugs)
	return bugData
}

func (orig *BugData) FilterByTargetRelease(sTargets []string) *BugData {
	bd := orig.clone()
	bugs := orig.GetBugs()

	targets := sets.NewString(sTargets...)
	filtered := []*bugzilla.Bug{}
	for i := range bugs {
		bug := bugs[i]
		if !targets.Has(bug.TargetRelease[0]) {
			continue
		}
		filtered = append(filtered, bug)
	}
	bd.set(filtered)
	return bd
}

func (orig *BugData) FilterBySeverity(sSeverities []string) *BugData {
	bd := orig.clone()
	bugs := bd.GetBugs()

	severities := sets.NewString(sSeverities...)
	filtered := []*bugzilla.Bug{}
	for i := range bugs {
		bug := bugs[i]
		if !severities.Has(bug.Severity) {
			continue
		}
		filtered = append(filtered, bug)
	}
	bd.set(filtered)
	return bd
}

// FIXME these should be loaded from orgData, not hard coded
func (orig *BugData) FilterBlocker() *BugData {
	bd := orig.FilterBySeverity([]string{"medium", "high", "urgent", "unspecified"})
	bd = bd.FilterByTargetRelease([]string{"---", CurrentRelease})
	return bd
}

func (orig *BugData) FilterByTeams(teams []string) *BugData {
	bd := orig.clone()
	teamMap := bd.GetTeamMap()
	bugs := []*bugzilla.Bug{}
	for _, team := range teams {
		bugs = append(bugs, teamMap[team]...)
	}
	bd.set(bugs)
	return bd
}

func (bd *BugData) GetBugs() []*bugzilla.Bug {
	bd.RLock()
	defer bd.RUnlock()
	return bd.bugs
}

func (bd *BugData) GetTeamMap() TeamMap {
	bugs := bd.GetBugs()
	teamMap := buildTeamMap(bugs, bd.orgData)
	return teamMap
}

func (bd *BugData) GetPeopleMap() PeopleMap {
	bugs := bd.GetBugs()
	teamMap := buildPeopleMap(bugs)
	return teamMap
}

func (bd *BugData) Length() int {
	bugs := bd.GetBugs()
	return len(bugs)
}

func (bd *BugData) set(bugs []*bugzilla.Bug) {
	bd.Lock()
	defer bd.Unlock()
	bd.bugs = bugs
}

func (bd *BugData) Reconcile() error {
	bugs, err := bd.client.Search(bd.query)
	if err != nil {
		return err
	}
	bd.set(bugs)
	return nil
}

type testClient struct {
	path string
}

func (tc testClient) UpdateBug(_ int, _ bugzilla.BugUpdate) error {
	return nil
}
func (tc testClient) Search(_ bugzilla.Query) ([]*bugzilla.Bug, error) {
	return []*bugzilla.Bug{}, nil
}
func (tc testClient) GetExternalBugPRsOnBug(_ int) ([]bugzilla.ExternalBug, error) {
	return []bugzilla.ExternalBug{}, nil
}
func (tc testClient) GetBug(_ int) (*bugzilla.Bug, error) {
	return &bugzilla.Bug{}, nil
}
func (tc testClient) Endpoint() string {
	return tc.path
}
func (testClient) AddPullRequestAsExternalBug(_ int, _ string, _ string, _ int) (bool, error) {
	return false, nil
}

func BugzillaClient(cmd *cobra.Command) (bugzilla.Client, error) {
	if testPath, err := cmd.Flags().GetString(bugDataFlagName); err != nil {
		return nil, err
	} else if testPath != "" {
		return bugzilla.GetTestClient(testPath), nil
	}

	endpoint := "https://bugzilla.redhat.com"

	keyFile, err := cmd.Flags().GetString(APIKeyFlagName)
	dat, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	apikey := strings.TrimRight(string(dat), "\r\n")

	var generator *func() []byte
	generatorFunc := func() []byte {
		return []byte(apikey)
	}
	generator = &generatorFunc

	return bugzilla.NewClient(*generator, endpoint), nil
}

func getAllOpenBugsQuery() bugzilla.Query {
	return bugzilla.Query{
		Classification: []string{"Red Hat"},
		Product:        []string{"OpenShift Container Platform"},
		Status:         []string{"NEW", "ASSIGNED", "POST", "ON_DEV", "MODIFIED"},
		IncludeFields:  []string{"id", "summary", "status", "severity", "priority", "assigned_to", "target_release", "component", "sub_components", "keywords"},
		Advanced: []bugzilla.AdvancedQuery{
			{
				Field:  "component",
				Op:     "equals",
				Value:  "Documentation",
				Negate: true,
			},
		},
	}
}

func buildPeopleMap(bugs []*bugzilla.Bug) PeopleMap {
	out := PeopleMap{}
	for i := range bugs {
		bug := bugs[i]
		assignee := bug.AssignedTo
		out[assignee] = append(out[assignee], bug)
	}

	return out
}

func buildTeamMap(bugs []*bugzilla.Bug, orgData *teams.OrgData) TeamMap {
	out := TeamMap{}
	for _, team := range orgData.Teams {
		out[team.Name] = []*bugzilla.Bug{}
	}
	out["unknown"] = []*bugzilla.Bug{}

	for i := range bugs {
		bug := bugs[i]
		team := orgData.GetTeam(bug)
		out[team] = append(out[team], bug)
	}

	return out
}

func getBugzillaAccess(cmd *cobra.Command) (bugzilla.Client, bugzilla.Query, error) {
	query := bugzilla.Query{}
	client, err := BugzillaClient(cmd)
	if err != nil {
		return client, query, err
	}
	query = getAllOpenBugsQuery()
	return client, query, nil
}

func (bd *BugData) Reconciler(errs chan error) {
	go func() {
		for true {
			if err := bd.Reconcile(); err != nil {
				errs <- err
				return
			}
			fmt.Printf("Successfully reconciled GetBugData. Teams:%d BugCount:%d\n", len(bd.orgData.Teams), len(bd.GetBugs()))
			time.Sleep(time.Minute * 5)
		}
	}()
}

func GetBugData(cmd *cobra.Command, orgData *teams.OrgData) (*BugData, error) {
	client, query, err := getBugzillaAccess(cmd)
	if err != nil {
		return nil, err
	}
	bugData := &BugData{
		cmd:     cmd,
		client:  client,
		query:   query,
		orgData: orgData,
	}
	err = bugData.Reconcile()
	if err != nil {
		return nil, err
	}
	return bugData, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(bugDataFlagName, bugDataFlagDefVal, bugDataFlagUsage)
	cmd.Flags().String(APIKeyFlagName, apiKeyFlagDefVal, apiKeyFlagUsage)
}
