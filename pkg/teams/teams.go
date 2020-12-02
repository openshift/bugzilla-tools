package teams

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/bugzilla-tools/pkg/config"

	"github.com/eparis/bugzilla"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	"github.com/imdario/mergo"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const (
	defaultForSubcomponentsTag = "!!DEFAULT!!"
	ignoreTeamTag              = "!!IGNORE!!"

	ignoreTeam
	fromGithubFlagName = "data-from-github"

	githubKeyFlagName   = "github-key"
	githubKeyFlagDefVal = "githubKey"

	teamDataFlagName   = "test-team-data"
	teamDataFlagDefVal = ""

	teamOverwriteFlagName   = "overwrite-team-data"
	teamOverwriteFlagDefVal = ""

	gsheetKeyFlagName   = "google-sheet"
	gsheetKeyFlagDefVal = "./"

	orgDataURLFlagName   = "org-data-url"
	orgDataURLFlagDefVal = "http://team-exportor/teams"
)

func isForTeam(team TeamInfo, componentToFind string, subcomponentToFind string) (isTeam, isDef bool) {
	foundComponent := false
	for _, component := range team.Components {
		if componentToFind == component {
			foundComponent = true
			break
		}
	}
	if !foundComponent {
		return false, false
	}
	subcomponents, ok := team.Subcomponents[componentToFind]
	if !ok {
		// Team has components, but no subcomponents, so all match
		return true, false
	}
	if len(subcomponents) == 1 && subcomponents[0] == defaultForSubcomponentsTag {
		return false, true
	}
	for _, subcomponent := range subcomponents {
		if subcomponentToFind == subcomponent {
			// both the component and the subcomponent match
			return true, false
		}
	}
	// Nothing matches
	return false, false
}

func (orgData OrgData) GetTeamByComponent(component, subcomponent string) *TeamInfo {
	var defTeam *TeamInfo
	for i := range orgData.Teams {
		team := orgData.Teams[i]
		yes, isDef := isForTeam(team, component, subcomponent)
		if yes {
			return &team
		}
		if isDef {
			defTeam = &team
		}
	}
	return defTeam
}

func (orgData OrgData) GetTeamName(bug *bugzilla.Bug) string {
	component := bug.Component[0]
	subcomponent := ""
	if subcomponents, ok := bug.SubComponent[component]; ok {
		subcomponent = subcomponents[0]
	}
	teamInfo := orgData.GetTeamByComponent(component, subcomponent)
	if teamInfo == nil {
		return "unknown"
	}
	return teamInfo.Name
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
func diskDataToOrgData(teamData DiskOrgData) (*OrgData, error) {
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
	orgData.SLO = teamData.SLO
	return orgData, nil
}

func GetGithubAuthClient(ctx context.Context, cmd *cobra.Command) (*http.Client, error) {

	apikey, err := config.GetConfigString(cmd, githubKeyFlagName, ctx)
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apikey},
	)
	tc := oauth2.NewClient(ctx, ts)
	return tc, nil
}

func getOrgDataFromLiPath(cmd *cobra.Command, path string) (*OrgData, error) {
	ctx := context.Background()
	transport, err := GetGithubAuthClient(ctx, cmd)
	if err != nil {
		return nil, err
	}
	client := github.NewClient(transport)
	file, _, _, err := client.Repositories.GetContents(ctx, "openshift", "li", path, nil)
	if err != nil {
		return nil, err
	}
	encoded := *file.Content
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	teamData := DiskOrgData{}
	err = yaml.Unmarshal(contents, &teamData)
	if err != nil {
		return nil, err
	}

	orgData, err := diskDataToOrgData(teamData)
	if err != nil {
		return nil, err
	}
	return orgData, nil
}

func getOrgDataFromFile(cmd *cobra.Command, whichFlag string) (*OrgData, error) {
	ctx := context.Background()
	teamData := DiskOrgData{}
	err := config.GetConfig(cmd, whichFlag, ctx, &teamData)
	if err != nil {
		return nil, err
	}

	orgData, err := diskDataToOrgData(teamData)
	if err != nil {
		return nil, err
	}
	return orgData, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(dir string, config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := filepath.Join(dir, "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getTeamSizeFromGoogleDoc(cmd *cobra.Command, orgData *OrgData) error {
	dir, err := cmd.Flags().GetString(gsheetKeyFlagName)
	if err != nil {
		return err
	}
	filename := filepath.Join(dir, "config.json")
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(dir, config)

	service, err := sheets.New(client)
	if err != nil {
		return err
	}
	readRange := `'OCP Team Structure'!B:C`
	resp, err := service.Spreadsheets.Values.Get("1M4C41fX2J1nBXhqPdtwd8UP4RAx98NA4ByIUv-0Z0Ds", readRange).Do()
	if err != nil {
		return err
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("No data found in google sheet")
	}

	data := map[string]int{}
	for _, row := range resp.Values {
		if len(row) != 2 {
			continue
		}
		team := row[0].(string)
		sizeStr := row[1].(string)
		size, err := strconv.ParseFloat(sizeStr, 32)
		if err != nil {
			continue
		}
		data[team] = int(size)
	}

	for team, count := range data {
		if teamInfo, ok := orgData.Teams[team]; ok {
			teamInfo.MemberCount = count
			orgData.Teams[team] = teamInfo
		} else {
			if team == "" {
				continue
			}
			pretty.Printf("Team %q: Found in Team Member Tracking Google Doc but not found in shiftzilla team config.\n", team)
		}
	}

	return nil
}

// This fetches org data from the github.com/openshift/li/tools/shiftzilla repo. That repo has most data
// inside shiftzilla_cfg.yaml, but shiftzilla doesn't understand subcomponents. So we have a second set
// of data which overwrites the first set and includes information about teams with subcomponents in bugzilla.
func getOrgDataFromLi(cmd *cobra.Command) (*OrgData, error) {
	primaryOrgData, err := getOrgDataFromLiPath(cmd, "tools/shiftzilla/shiftzilla_cfg.yaml")
	if err != nil {
		return nil, err
	}

	secondaryOrgData, err := getOrgDataFromLiPath(cmd, "tools/shiftzilla/subcomponent_teams.yaml")
	if err != nil {
		return nil, err
	}

	if err = mergo.MergeWithOverwrite(primaryOrgData, secondaryOrgData); err != nil {
		return nil, err
	}
	return primaryOrgData, nil
}

// This could actually get the data from local file (--test-data=) or from github itself.
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

	err = getTeamSizeFromGoogleDoc(cmd, orgData)
	if err != nil {
		return nil, err
	}

	toIgnore := []string{}
	for teamName, team := range orgData.Teams {
		if len(team.Components) == 1 && team.Components[0] == ignoreTeamTag {
			toIgnore = append(toIgnore, teamName)
		}
	}
	for _, teamName := range toIgnore {
		fmt.Printf("Deleting: %s\n", teamName)
		delete(orgData.Teams, teamName)
	}

	orgData.cmd = cmd
	return orgData, nil
}

func getOrgDataFromService(cmd *cobra.Command) (*OrgData, error) {
	url, err := cmd.Flags().GetString(orgDataURLFlagName)
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

// CurrentVersion returns the lowest x.y version that has a x.y.0 target.
func (orgData OrgData) CurrentVersion() (string, error) {
	var all []string
	var active []string
	var order []string
	for _, v := range orgData.Releases {
		all = append(all, v.Name)
		onlyZ := true
		for _, target := range v.Targets {
			if !strings.HasSuffix(target, ".z") {
				onlyZ = false
				break
			}
		}
		if onlyZ {
			continue
		}

		vs := strings.Split(v.Name, ".")
		if len(vs) < 2 {
			continue
		}
		active = append(active, fmt.Sprintf("%s.%s", vs[0], vs[1]))
		if len(vs[1]) == 1 {
			vs[1] = "0" + vs[1]
		}
		order = append(order, vs[0]+vs[1])
	}
	if len(active) == 0 {
		return "", fmt.Errorf("no release found that has a x.y.0 target version: %v", all)
	}
	sort.Slice(active, func(i, j int) bool {
		return order[i] < order[j]
	})
	return active[0], nil
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
	cmd.Flags().String(gsheetKeyFlagName, gsheetKeyFlagDefVal, "Path to file containing google sheets oauth keys")
	cmd.Flags().String(orgDataURLFlagName, orgDataURLFlagDefVal, "URL to Load Org Data, http://localhost:8000/teams might be your choice locally")
}
