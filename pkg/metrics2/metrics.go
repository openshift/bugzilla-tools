package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func updateGauge(bugs bugs.BugMap, bugGauge *prometheus.GaugeVec) {
	for team, bugs := range bugs {
		for _, bug := range bugs {
			label := prometheus.Labels{
				"team":           team,
				"id":             fmt.Sprintf("%d", bug.ID),
				"status":         bug.Status,
				"severity":       bug.Severity,
				"keywords":       strings.Join(bug.Keywords, ","),
				"target_release": bug.TargetRelease[0],
			}
			bugGauge.With(label).Set(1)
		}
	}
}

func createGauge() *prometheus.GaugeVec {
	ops := prometheus.GaugeOpts{
		Name: "bugs",
		Help: "All Bugs",
	}
	requiredLabels := []string{"team", "id", "status", "severity", "keywords", "target_release"}
	return promauto.NewGaugeVec(ops, requiredLabels)
}

// Create a guague for every team.
func Setup(errs chan error, bugData *bugs.BugData) {
	bugGauge := createGauge()
	go func() {
		for true {
			bugs := bugData.GetBugMap()
			fmt.Printf("Found %d teams in bugMap!\n", len(bugs))
			// Don't publish data until we actually get a response from BZ
			if len(bugs) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}
			updateGauge(bugs, bugGauge)
			time.Sleep(1 * time.Minute)
		}
	}()
}
