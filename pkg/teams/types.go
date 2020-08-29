package teams

import (
	"github.com/spf13/cobra"

	sloAPI "github.com/openshift/bugzilla-tools/pkg/slo/api"
)

const (
	reconcileService = "service"
	reconcileFiles   = "files"
)

type TeamInfo struct {
	Name          string                 `json:"name,omitempty"`
	SlackChan     string                 `json:"slack_chan,omitempty"`
	Lead          string                 `json:"lead,omitempty"`
	Managers      []string               `json:"managers,omitempty"`
	Group         string                 `json:"group,omitempty"`
	Components    []string               `json:"components,omitempty"`
	Subcomponents map[string][]string    `json:"subcomponents,omitempty"`
	MemberCount   int                    `json:"memberCount,omitempty"`
	SLO           map[string]sloAPI.Data `json:"slo,omitempty"`
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

type DiskOrgData struct {
	OrgTitle string                 `json:"OrgTitle,omitempty"`
	Teams    []TeamInfo             `json:"Teams,omitempty"`
	Releases []ReleaseInfo          `json:"Releases,omitempty"`
	SLO      map[string]sloAPI.Data `json:"slo,omitempty"`
}

type OrgData struct {
	OrgTitle string                 `json:"orgTitle,omitempty"`
	Teams    map[string]TeamInfo    `json:"teams,omitempty"`
	Releases map[string]ReleaseInfo `json:"releases,omitempty"`
	SLO      map[string]sloAPI.Data `json:"slo,omitempty"`
	cmd      *cobra.Command
}
