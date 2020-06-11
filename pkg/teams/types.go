package teams

import (
	"github.com/spf13/cobra"
)

type TeamInfo struct {
	Name          string              `json:"name,omitempty"`
	Lead          string              `json:"lead,omitempty"`
	Managers      []string            `json:"managers,omitempty"`
	Group         string              `json:"group,omitempty"`
	Components    []string            `json:"components,omitempty"`
	Subcomponents map[string][]string `json:"subcomponents,omitempty"`
}

type Milestones struct {
	Start           string `json:"start,omitempty"`
	FeatureComplete string `json:"feature_complete,omitempty"`
	CodeFreeze      string `json:"code_freeze,omitempty"`
	GA              string `json:"ga,omitempty"`
}

type ReleaseInfo struct {
	Name       string      `json:"name,omitempty"`
	Targets    []string    `json:"targets,omitempty"`
	Milestones *Milestones `json:"milestones,omitempty"`
}

type Teams struct {
	OrgTitle string        `json:"OrgTitle,omitempty"`
	Teams    []TeamInfo    `json:"Teams,omitempty"`
	Releases []ReleaseInfo `json:"Releases,omitempty"`
}

type OrgData struct {
	OrgTitle string                 `json:"orgTitle,omitempty"`
	Teams    map[string]TeamInfo    `json:"teams,omitempty"`
	Releases map[string]ReleaseInfo `json:"releases,omitempty"`
	cmd      *cobra.Command
}
