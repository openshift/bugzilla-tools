package metrics

import (
	"time"

	"github.com/eparis/react-material/pkg/bugs"
	"github.com/eparis/react-material/pkg/teams"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type bugMetrics struct {
	all                     *prometheus.GaugeVec
	upcomingSprint          *prometheus.GaugeVec
	mediumAndHigherSeverity *prometheus.GaugeVec
}

func updateGauge(teamName string, countFunc func(string) int, gauge *prometheus.GaugeVec) {
	label := prometheus.Labels{
		"team": teamName,
	}

	count := float64(countFunc(teamName))
	gauge.With(label).Set(count)
}

func updateGauges(teamName string, bugs bugs.BugMap, bugMetrics bugMetrics) {
	updateGauge(teamName, bugs.CountAll, bugMetrics.all)
	updateGauge(teamName, bugs.CountUpcomingSprint, bugMetrics.upcomingSprint)
	updateGauge(teamName, bugs.CountNotLowSeverity, bugMetrics.mediumAndHigherSeverity)
}

func updateCounts(teams teams.Teams, bugs bugs.BugMap, bugMetrics bugMetrics) {
	for i := range teams.Teams {
		teamName := teams.Teams[i].Name
		updateGauges(teamName, bugs, bugMetrics)
	}
	updateGauges("unknown", bugs, bugMetrics)
}

func createGauge(name, help string) *prometheus.GaugeVec {
	ops := prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}
	return promauto.NewGaugeVec(ops, []string{"team"})
}

func createGauges() bugMetrics {
	bugMetrics := bugMetrics{}

	name := "bugs"
	help := "Total bugs"
	bugMetrics.all = createGauge(name, help)

	name = "bugs_with_upcoming_sprint"
	help = "Number of bugs not marked 'UpcomingSprint'"
	bugMetrics.upcomingSprint = createGauge(name, help)

	name = "bugs_medium_and_higher_severity"
	help = "Number of medium or higher severity bugs"
	bugMetrics.mediumAndHigherSeverity = createGauge(name, help)
	return bugMetrics
}

// Create a guague for every team.
func Setup(errs chan error, bugData *bugs.BugData, teams teams.Teams) {
	bugMetrics := createGauges()
	go func() {
		for true {
			bugs := bugData.GetBugMap()
			// Don't publish data until we actually get a response from BZ
			if len(bugs) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}
			updateCounts(teams, bugs, bugMetrics)
			time.Sleep(1 * time.Minute)
		}
	}()
}
