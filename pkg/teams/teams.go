package teams

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	githubKeyFlagName   = "github-key"
	githubKeyFlagDefVal = "githubKey"

	teamDataFlagName   = "test-team-data"
	teamDataFlagDefVal = ""
)

func (teams Teams) GetTeam(componentToFind string) string {
	for _, team := range teams.Teams {
		for _, component := range team.Components {
			if componentToFind == component {
				return team.Name
			}
		}
	}
	return "unknown"
}

func GetTeamData(cmd *cobra.Command) (Teams, error) {
	teams := Teams{}
	if path, err := cmd.Flags().GetString(teamDataFlagName); err != nil {
		return teams, err
	} else if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return teams, err
		}
		err = yaml.Unmarshal(data, &teams)
		return teams, err
	}

	keyFile, err := cmd.Flags().GetString("github-key")
	dat, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return teams, err
	}
	apikey := strings.TrimRight(string(dat), "\r\n")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apikey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	file, _, _, err := client.Repositories.GetContents(ctx, "openshift", "li", "tools/shiftzilla/shiftzilla_cfg.yaml", nil)
	if err != nil {
		return teams, err
	}
	encoded := *file.Content
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return teams, err
	}

	err = yaml.Unmarshal(contents, &teams)
	if err != nil {
		return teams, err
	}
	return teams, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(githubKeyFlagName, githubKeyFlagDefVal, "Path to file containing github key")
	cmd.Flags().String(teamDataFlagName, teamDataFlagDefVal, "Path to file containing team data")
}
