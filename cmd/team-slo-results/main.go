package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	//"github.com/kr/pretty"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/slo"
	sloAPI "github.com/openshift/bugzilla-tools/pkg/slo/api"
	"github.com/openshift/bugzilla-tools/pkg/teams"
)

const (
	port = "8001"
)

func getTeamSLOResults(cmd *cobra.Command, orgInfo *teams.OrgData, bugData *bugs.BugData) (sloAPI.TeamsResults, error) {
	bugMaps := slo.GetBugMaps(bugData)

	currentVersion, err := orgInfo.CurrentVersion()
	if err != nil {
		return nil, err
	}

	// TODO: consider more releases?
	ciComponentMap, err := slo.GetCiComponentMap(currentVersion)
	if err != nil {
		return nil, err
	}

	teamsResults := make(sloAPI.TeamsResults, len(orgInfo.Teams))
	for team, teamInfo := range orgInfo.Teams {
		teamsResults[team] = slo.GetTeamResult(bugMaps, ciComponentMap, orgInfo, teamInfo)
	}
	return teamsResults, nil
}

func GetTeamHandler(data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Printf("Unable to encode: %v: %v", data, err)
		}
	}
}

func serveHTTP(errs chan error, serveResults *sloAPI.TeamsResults) {
	mux := http.NewServeMux()
	mux.Handle("/teams", GetTeamHandler(serveResults))
	mux.Handle("/metrics", promhttp.Handler())

	staticHandler := http.FileServer(http.Dir("./web/build/"))
	mux.Handle("/", staticHandler)

	listenAt := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    listenAt,
		Handler: mux,
	}

	go func() {
		errs <- srv.ListenAndServe()
	}()
}

func doMain(cmd *cobra.Command) error {
	errs := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	orgInfo, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}
	orgInfo.Reconciler()

	bugData, err := bugs.GetBugData(cmd, orgInfo)
	if err != nil {
		return err
	}
	bugData.Reconciler(errs)
	bugData = bugData.FilterByStatus(bugs.OnEngineeringStatus())

	serveResults := &sloAPI.TeamsResults{}

	go func() {
		for {
			teamsResults, err := getTeamSLOResults(cmd, orgInfo, bugData)
			if err != nil {
				errs <- err
				return

			}
			if len(teamsResults) == 0 {
				time.Sleep(2 * time.Second)
				continue
			}
			*serveResults = teamsResults
			time.Sleep(10 * time.Minute)
		}
	}()
	serveHTTP(errs, serveResults)

	fmt.Println("http server started.")

	err = nil
	select {
	case <-stop:
		fmt.Println("Sutting down...")
	case outErr := <-errs:
		fmt.Println("Failed to start server:", outErr.Error())
		err = outErr
	}
	return err
}

func main() {
	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := doMain(cmd)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
