package jira

import (
	"io/ioutil"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

const (
	endpoint = "https://issues.redhat.com"
	user     = "eparis"

	keyFlagName   = "jira-key"
	keyFlagDefVal = "jiraKey"
	keyFlagUsage  = "Path to file containing Jira basic auth password"

	issuePathFlagName   = "issue-path"
	issuePathFlagDefVal = "issues/"
	issuePathFlagUsage  = "Path to directory containing issue snapshots"
)

func GetIssues(client *jira.Client, searchString string) (map[string]jira.Issue, error) {
	var issues map[string]jira.Issue
	last := 0
	for {
		opt := &jira.SearchOptions{
			MaxResults: 100,
			StartAt:    last,
		}

		chunk, resp, err := client.Issue.Search(searchString, opt)
		if err != nil {
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make(map[string]jira.Issue, total)
		}
		for i := range chunk {
			issue := chunk[i]
			issues[issue.Key] = issue
		}
		last = resp.StartAt + len(chunk)
		if last >= total {
			break
		}
	}
	return issues, nil
}

func GetClient(cmd *cobra.Command) (*jira.Client, error) {
	keyFile, err := cmd.Flags().GetString(keyFlagName)
	dat, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	jiraPassword := strings.TrimRight(string(dat), "\r\n")
	_ = jiraPassword

	tp := jira.BasicAuthTransport{
		Username: user,
		Password: jiraPassword,
	}

	client, err := jira.NewClient(tp.Client(), endpoint)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(keyFlagName, keyFlagDefVal, keyFlagUsage)
	cmd.Flags().String(issuePathFlagName, issuePathFlagDefVal, issuePathFlagUsage)
}
