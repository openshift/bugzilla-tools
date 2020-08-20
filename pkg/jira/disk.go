package jira

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

type Issues struct {
	Date   time.Time
	Issues map[string]jira.Issue
}

func issueDir(cmd *cobra.Command) string {
	dir, err := cmd.Flags().GetString(issuePathFlagName)
	if err != nil {
		panic(err)
	}
	return dir
}

func WriteIssues(cmd *cobra.Command, issues map[string]jira.Issue) error {
	now := time.Now()
	date := now.Format("2006-01-02")
	filename := date + ".yaml"
	dir := issueDir(cmd)
	path := filepath.Join(dir, filename)

	dateIssues := Issues{
		Date:   now,
		Issues: issues,
	}

	out, err := yaml.Marshal(&dateIssues)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(out), 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetSnapshots(cmd *cobra.Command) ([]string, error) {
	var files []string

	dir := issueDir(cmd)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		path = filepath.Base(path)
		path = strings.TrimSuffix(path, ".yaml")
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func GetDiskIssues(cmd *cobra.Command, date string) (map[string]jira.Issue, error) {
	filename := fmt.Sprintf("%s.yaml", date)
	dir := issueDir(cmd)
	path := filepath.Join(dir, filename)

	dateIssues := Issues{}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &dateIssues)
	if err != nil {
		return nil, err
	}

	return dateIssues.Issues, nil
}
