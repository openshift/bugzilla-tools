package teams

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/openshift/bugzilla-tools/pkg/config"

	"github.com/eparis/bugzilla"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	"github.com/imdario/mergo"
	//"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	fromGithubFlagName = "data-from-github"

	githubKeyFlagName   = "github-key"
	githubKeyFlagDefVal = "githubKey"

	teamDataFlagName   = "test-team-data"
	teamDataFlagDefVal = ""

	teamOverwriteFlagName   = "overwrite-team-data"
	teamOverwriteFlagDefVal = ""
)

func isForTeam(team TeamInfo, componentToFind string, subcomponentToFind string) bool {
	foundComponent := false
	for _, component := range team.Components {
		if componentToFind == component {
			foundComponent = true
			break
		}
	}
	if !foundComponent {
		return false
	}
	subcomponents, ok := team.Subcomponents[componentToFind]
	if !ok {
		// Team has components, but no subcomponents, so all match
		return true
	}
	for _, subcomponent := range subcomponents {
		if subcomponentToFind == subcomponent {
			// both the component and the subcomponent match
			return true
		}
	}
	// Nothing matches
	return false
}

func (orgData OrgData) GetTeam(bug *bugzilla.Bug) string {
	component := bug.Component[0]
	subcomponent := ""
	if subcomponents, ok := bug.SubComponent[component]; ok {
		subcomponent = subcomponents[0]
	}

	for _, team := range orgData.Teams {
		if isForTeam(team, component, subcomponent) {
			return team.Name
		}
	}
	return "unknown"
}

func (orgData OrgData) GetTeamNames() []string {
	out := make([]string, 0, len(orgData.Teams))
	for team, _ := range orgData.Teams {
		out = append(out, team)
	}
	sort.Strings(out)
	return out
}

// mainly move from the list of teams and releases to a map[name]team or map[name]release
func teamDataToOrgData(teamData Teams) (*OrgData, error) {
	orgData := &OrgData{}
	orgData.OrgTitle = teamData.OrgTitle
	orgData.Teams = map[string]TeamInfo{}
	for i := range teamData.Teams {
		teamInfo := teamData.Teams[i]
		name := teamInfo.Name
		orgData.Teams[name] = teamInfo
	}
	orgData.Releases = map[string]ReleaseInfo{}
	for i := range teamData.Releases {
		releaseInfo := teamData.Releases[i]
		name := releaseInfo.Name
		orgData.Releases[name] = releaseInfo
	}
	return orgData, nil
}

func getOrgDataFromLi(cmd *cobra.Command) (*OrgData, error) {
	ctx := context.Background()

	apikey, err := config.GetConfigString(cmd, githubKeyFlagName, ctx)
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apikey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	file, _, _, err := client.Repositories.GetContents(ctx, "openshift", "li", "tools/shiftzilla/shiftzilla_cfg.yaml", nil)
	if err != nil {
		return nil, err
	}
	encoded := *file.Content
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	teamData := Teams{}
	err = yaml.Unmarshal(contents, &teamData)
	if err != nil {
		return nil, err
	}

	orgData, err := teamDataToOrgData(teamData)
	if err != nil {
		return nil, err
	}
	return orgData, nil
}

func getOrgDataFromFile(cmd *cobra.Command, whichFlag string) (*OrgData, error) {
	ctx := context.Background()
	teamData := Teams{}
	err := config.GetConfig(cmd, whichFlag, ctx, &teamData)
	if err != nil {
		return nil, err
	}

	orgData, err := teamDataToOrgData(teamData)
	if err != nil {
		return nil, err
	}
	return orgData, nil
}

func getOrgDataFromGithub(cmd *cobra.Command) (*OrgData, error) {
	orgData, err := getOrgDataFromFile(cmd, teamDataFlagName)
	if err != nil && err != config.NotSetError {
		// bail if we got a real error
		return nil, err
	} else if err == config.NotSetError {
		// if the error was that the flag wasn't set pull from github
		orgData, err = getOrgDataFromLi(cmd)
		if err != nil {
			return nil, err
		}
	}

	// get the overwrite data
	overrideOrgData, err := getOrgDataFromFile(cmd, teamOverwriteFlagName)
	if err != nil && err != config.NotSetError {
		return nil, err
	} else if err == nil {
		// merge overwrite with the main data
		if err = mergo.MergeWithOverwrite(orgData, overrideOrgData); err != nil {
			return nil, err
		}
	}
	orgData.cmd = cmd
	return orgData, nil
}

func getOrgDataFromService(cmd *cobra.Command) (*OrgData, error) {
	url := "http://team-exportor/teams"

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "github.com/openshift/bugzilla-tools/pkg/teams")

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

	orgData := &OrgData{}
	jsonErr := json.Unmarshal(body, orgData)
	if jsonErr != nil {
		return nil, err
	}
	orgData.cmd = cmd

	return orgData, nil
}

func getOrgData(cmd *cobra.Command) (*OrgData, error) {
	fromGithub, err := cmd.Flags().GetBool(fromGithubFlagName)
	if err != nil {
		return nil, err
	}
	var orgData *OrgData
	if fromGithub {
		orgData, err = getOrgDataFromGithub(cmd)
	} else {
		orgData, err = getOrgDataFromService(cmd)
	}
	return orgData, err
}

func GetOrgData(cmd *cobra.Command) (*OrgData, error) {
	return getOrgData(cmd)
}

func (orgData *OrgData) Reconcile() {
	newOrgData, err := getOrgData(orgData.cmd)
	if err != nil {
		log.Fatalln(err)
	}
	*orgData = *newOrgData
}

func (orgData *OrgData) Reconciler() {
	go func() {
		for true {
			orgData.Reconcile()
			fmt.Printf("Successfully fetched OrgData len(teams):%d len(releases): %d\n", len(orgData.Teams), len(orgData.Releases))
			time.Sleep(time.Minute * 5)
		}
	}()
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(fromGithubFlagName, false, "Use github and local files or use the microservice")
	cmd.Flags().String(githubKeyFlagName, githubKeyFlagDefVal, "Path to file containing github key")
	cmd.Flags().String(teamDataFlagName, teamDataFlagDefVal, "Path to file containing team data")
	cmd.Flags().String(teamOverwriteFlagName, teamOverwriteFlagDefVal, "Path to file containing team data to overwrite with github/file data")
}
