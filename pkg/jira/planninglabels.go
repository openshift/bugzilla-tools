package jira

import (
	"github.com/andygrunwald/go-jira"
	"github.com/ghodss/yaml"
	"github.com/kr/pretty"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	planningFieldKey = "customfield_12316343"
)

type OpenShiftPlanningLabel map[string]string

type OpenShiftPlanning []OpenShiftPlanningLabel

func junk() {
	pretty.Println("junk")
}

func convert(in interface{}, out interface{}) error {
	b, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, out)
	if err != nil {
		return err
	}
	return nil
}

func GetPlanningLabels(issue *jira.Issue) (sets.String, error) {
	out := sets.NewString()

	planningField := issue.Fields.Unknowns[planningFieldKey]
	openshiftPlanning := OpenShiftPlanning{}
	err := convert(planningField, &openshiftPlanning)
	if err != nil {
		return nil, err
	}
	for _, label := range openshiftPlanning {
		value := label["value"]
		out.Insert(value)
	}
	return out, nil
}
