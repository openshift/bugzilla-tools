package sloAPI

const (
	Urgent              = "urgents"
	Blocker             = "blockers"
	UrgentCustomerCases = "urgent-customer-cases"
	All                 = "total"
	PMScore             = "pmscore"
	CI                  = "ci-fail-rate"
)

var (
	OrderedSLOs = []string{
		Urgent,
		UrgentCustomerCases,
		Blocker,
		PMScore,
		CI,
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
