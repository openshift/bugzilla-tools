package slo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/kr/pretty"
	"github.com/spf13/cobra"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	sloAPI "github.com/openshift/bugzilla-tools/pkg/slo/api"
	"github.com/openshift/bugzilla-tools/pkg/teams"
)

const (
	sloResultsURLFlagName   = "slo-results-url"
	sloResultsURLFlagDefVal = "http://team-slo-resluts/teams"
)

func GetTeamsResults(cmd *cobra.Command) (*sloAPI.TeamsResults, error) {
	url, err := cmd.Flags().GetString(sloResultsURLFlagName)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "github.com/openshift/bugzilla-tools/pkg/slo")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	teamsResults := &sloAPI.TeamsResults{}
	jsonErr := json.Unmarshal(body, teamsResults)
	if jsonErr != nil {
		return nil, err
	}

	return teamsResults, nil
}

func GetBugMaps(bugData *bugs.BugData) map[string]bugs.TeamMap {
	bugMaps := map[string]bugs.TeamMap{
		sloAPI.All:     bugData.GetTeamMap(),
		sloAPI.Urgent:  bugData.FilterBySeverity([]string{"urgent"}).GetTeamMap(),
		sloAPI.Blocker: bugData.FilterBlocker().GetTeamMap(),
	}
	return bugMaps
}

func getCountResult(which string, bugMaps map[string]bugs.TeamMap, teamSLO map[string]sloAPI.Data, teamInfo teams.TeamInfo) sloAPI.Result {
	team := teamInfo.Name
	sloData := teamSLO[which]
	bugs, ok := bugMaps[which]
	if !ok {
		pretty.Printf("Unable to find bug map for SLO: %s\n", which)
		return sloAPI.Result{}
	}
	current := len(bugs[team])
	obligation := sloData.Count
	if sloData.PerMember {
		obligation = obligation * teamInfo.MemberCount
	}
	result := sloAPI.Result{
		Name:       which,
		Current:    current,
		Obligation: obligation,
		PerMember:  sloData.PerMember,
	}
	return result
}

func getPMScoreResult(bugMap bugs.TeamMap, sloData sloAPI.Data, teamInfo teams.TeamInfo) sloAPI.Result {
	team := teamInfo.Name
	score := 0
	for _, bug := range bugMap[team] {
		bugScore, err := strconv.Atoi(bug.PMScore)
		if err != nil {
			bugScore = 1
		}
		score += bugScore
	}
	obligation := sloData.Count
	if sloData.PerMember {
		obligation = obligation * teamInfo.MemberCount
	}
	result := sloAPI.Result{
		Name:       sloAPI.PMScore,
		Current:    score,
		Obligation: obligation,
		PerMember:  sloData.PerMember,
	}
	return result
}

func GetTeamResult(bugMaps map[string]bugs.TeamMap, orgData *teams.OrgData, teamInfo teams.TeamInfo) sloAPI.TeamResult {
	team := teamInfo.Name
	if teamInfo.MemberCount == 0 {
		pretty.Printf("%s has 0 members\n", team)
	}

	teamResult := sloAPI.TeamResult{
		Name:    team,
		Members: teamInfo.MemberCount,
	}

	teamSLO := make(map[string]sloAPI.Data, len(orgData.SLO))
	// Set the org wide SLOs
	for key, value := range orgData.SLO {
		teamSLO[key] = value
	}
	// Override with the team specific SLOs
	for key, value := range teamInfo.SLO {
		teamSLO[key] = value
	}
	// Check them all
	for _, which := range sloAPI.OrderedSLOs {
		var result sloAPI.Result
		switch which {
		case sloAPI.All:
			result = getCountResult(which, bugMaps, teamSLO, teamInfo)
		case sloAPI.Urgent:
			result = getCountResult(which, bugMaps, teamSLO, teamInfo)
		case sloAPI.Blocker:
			result = getCountResult(which, bugMaps, teamSLO, teamInfo)
		case sloAPI.PMScore:
			bugMap := bugMaps[sloAPI.All]
			sloData := teamSLO[sloAPI.PMScore]
			result = getPMScoreResult(bugMap, sloData, teamInfo)
		default:
			panic("Didn't know an SLO!!!")
		}
		if result.Current > result.Obligation {
			teamResult.Failing = true
		}
		teamResult.Results = append(teamResult.Results, result)
	}
	return teamResult
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(sloResultsURLFlagName, sloResultsURLFlagDefVal, "URL to the SLO Results. http://localhost:8001/teams is a good choice for running locally")
}
