package teams

import ()

type TeamInfo struct {
	Name       string   `json:"name,omitempty"`
	Lead       string   `json:"lead,omitempty"`
	Group      string   `json:"group,omitempty"`
	Components []string `json:"components,omitempty"`
}

type Milestones struct {
	Start           string `json:"start,omitempty"`
	FeatureComplete string `json:"feature_complete,omitempty"`
	CodeFreeze      string `json:"code_freeze,omitempty"`
	GA              string `json:"ga,omitempty"`
}

type ReleaseInfo struct {
	Name       string     `json:"name, omitempty"`
	Targets    []string   `json:"targets,omitempty"`
	Milestones Milestones `json:"milestones,omitempty"`
}

type Teams struct {
	OrgTitle string
	Teams    []TeamInfo
	Releases []ReleaseInfo
}
