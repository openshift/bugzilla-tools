package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/slo"
	"github.com/openshift/bugzilla-tools/pkg/teams"
	"github.com/openshift/bugzilla-tools/pkg/utils"
)

const (
	//githubQuery = `repo:openshift/ovn-kubernetes label:lgtm label:approved is:open is:pr base:master base:main`
	githubQuery = `org:openshift label:lgtm label:approved is:open is:pr base:master base:main`
)

type RepoToBugzillaInfo struct {
	Component    string `json:"component,omitempty"`
	Subcomponent string `json:"subcomponent,omitempty"`
}

func GetRepoFromIssue(ctx context.Context, client *github.Client, issue *github.Issue) (*github.Repository, error) {
	u := issue.RepositoryURL
	req, err := client.NewRequest("GET", *u, nil)
	if err != nil {
		return nil, err
	}

	repository := new(github.Repository)
	_, err = client.Do(ctx, req, repository)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func hasValidBug(issue *github.Issue) bool {
	labelSet := sets.NewString()
	for _, label := range issue.Labels {
		labelSet.Insert(label.GetName())
	}
	if labelSet.Has("bugzilla/valid-bug") {
		return true
	}
	return false
}

func doBug(cmd *cobra.Command) error {
	ctx := context.TODO()

	orgInfo, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	teamsSLOResults, err := slo.GetTeamsResults(cmd)
	if err != nil {
		return err
	}

	authClient, err := teams.GetGithubAuthClient(ctx, cmd)
	if err != nil {
		return err
	}

	diskCache := diskcache.New("cache")
	transport := httpcache.NewTransport(diskCache)
	transport.Transport = authClient.Transport
	httpclient := &http.Client{
		Transport: transport,
	}
	githubClient := github.NewClient(httpclient)

	searchOpts := github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	var allIssues []*github.Issue
	for {
		issues, resp, err := githubClient.Search.Issues(ctx, githubQuery, &searchOpts)
		if err != nil {
			return err
		}

		if allIssues == nil {
			allIssues = make([]*github.Issue, 0, *issues.Total)
		}
		allIssues = append(allIssues, issues.Issues...)
		if resp.NextPage == 0 {
			break
		}
		searchOpts.Page = resp.NextPage

	}

	teamKnown := 0
	repoAllowed := map[string]int{}
	repoDenied := map[string]int{}
	componentUnknown := map[string]int{}
	for _, issue := range allIssues {
		repo, err := GetRepoFromIssue(ctx, githubClient, issue)
		if err != nil {
			return err
		}

		repoName := *repo.Name

		if hasValidBug(issue) {
			repoAllowed[repoName] = repoAllowed[repoName] + 1
			continue
		}

		org := repo.GetOrganization()
		orgName := org.Login
		file, _, _, err := githubClient.Repositories.GetContents(ctx, *orgName, repoName, "OWNERS", nil)
		if err != nil {
			return err
		}
		contents, err := file.GetContent()
		if err != nil {
			return err
		}

		bugInfo := RepoToBugzillaInfo{}
		err = yaml.Unmarshal([]byte(contents), &bugInfo)
		if err != nil {
			return err
		}

		teamInfo := orgInfo.GetTeamByComponent(bugInfo.Component, bugInfo.Subcomponent)
		if teamInfo == nil {
			componentUnknown[repoName] = componentUnknown[repoName] + 1
			continue
		}

		team := teamInfo.Name

		sloResults := (*teamsSLOResults)[team]
		teamKnown += 1

		if sloResults.Failing {
			repoDenied[repoName] = repoDenied[repoName] + 1
		} else {
			repoAllowed[repoName] = repoAllowed[repoName] + 1
		}
	}

	pretty.Println("*************ALLOWED*******************")
	for _, team := range utils.SortedKeys(repoAllowed) {
		pretty.Printf("%s:%d\n", team, repoAllowed[team])
	}
	pretty.Println()
	pretty.Println("*************DENIED*******************")
	for _, team := range utils.SortedKeys(repoDenied) {
		pretty.Printf("%s:%d\n", team, repoDenied[team])
	}

	pretty.Println()
	pretty.Println("*************COMPONENT UNKNOWN*******************")
	for _, team := range utils.SortedKeys(componentUnknown) {
		pretty.Printf("%s:%d\n", team, componentUnknown[team])
	}
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := doBug(cmd)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Flags().StringSlice("actions", []string{}, "Actions to run, unset runs all actions with default=true")
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)
	slo.AddFlags(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
