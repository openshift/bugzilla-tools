package jira

import (
	"github.com/andygrunwald/go-jira"
	"k8s.io/apimachinery/pkg/util/sets"
)

func diffIssue(oldIssue, newIssue jira.Issue) (string, error) {
	return "", nil
}

func DiffIssueLists(oldIssues, newIssues map[string]jira.Issue) (added []string, removed []string, err error) {
	oldSet := sets.NewString()
	for key := range oldIssues {
		oldSet.Insert(key)
	}

	newSet := sets.NewString()
	for key := range newIssues {
		newSet.Insert(key)
	}

	removed = oldSet.Difference(newSet).List()
	added = newSet.Difference(oldSet).List()

	return added, removed, nil
}

type IssueInfo struct {
	Summary        string   `json:"summary"`
	Status         string   `json:"status"`
	Key            string   `json:"key"`
	PlanningLabels []string `json:"planninglabels"`
	FixedVersions  []string `json:"fixedversions"`
}

func getFixVersions(issue *jira.Issue) sets.String {
	versions := sets.NewString()
	for _, version := range issue.Fields.FixVersions {
		versions.Insert(version.Name)
	}
	return versions
}

func getIssueInfo(issue *jira.Issue) (IssueInfo, error) {
	out := IssueInfo{
		Summary:       issue.Fields.Summary,
		Key:           issue.Key,
		FixedVersions: getFixVersions(issue).List(),
		Status:        issue.Fields.Status.Name,
	}
	planningLabels, err := GetPlanningLabels(issue)
	if err != nil {
		return out, err
	}
	out.PlanningLabels = planningLabels.List()
	return out, nil
}

func GetIssuesInfo(client *jira.Client, issues []string) ([]IssueInfo, error) {
	out := make([]IssueInfo, 0, len(issues))
	for _, key := range issues {
		issue, _, err := client.Issue.Get(key, nil)
		if err != nil {
			return nil, err
		}
		issueInfo, err := getIssueInfo(issue)
		if err != nil {
			return nil, err
		}
		out = append(out, issueInfo)
	}

	return out, nil
}

type JiraDiffData struct {
	NewDate string      `json:"newdate"`
	OldDate string      `json:"olddate"`
	Added   []IssueInfo `json:"added"`
	Removed []IssueInfo `json:"removed"`
}
