package sloAPI

const (
	Urgent  = "urgents"
	Blocker = "blockers"
	All     = "total"
	PMScore = "pmscore"
)

var (
	OrderedSLOs = []string{
		Urgent,
		Blocker,
		PMScore,
		All,
	}
)

type Result struct {
	Name       string `json:"name,omitempty"`
	Current    int    `json:"current"`
	Obligation int    `json:"obligation"`
	PerMember  bool   `json:"perMember,omitempty"`
}

type TeamResult struct {
	Name    string   `json:"name"`
	Failing bool     `json:"failing"`
	Members int      `json:"members,omitempty"`
	Results []Result `json:"results,omitempty"`
}

type TeamsResults map[string]TeamResult

type Data struct {
	Count     float32 `json:"count,omitempty"`
	PerMember bool    `json:"perMember,omitempty"`
}
