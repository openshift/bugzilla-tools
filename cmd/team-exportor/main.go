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

	"github.com/eparis/bugtool/pkg/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	port = "8000"
)

func GetTeamHandler(orgData *teams.OrgData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams := orgData.Teams
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewEncoder(w).Encode(teams)
		if err != nil {
			fmt.Errorf("Unable to encode: %v: %v", teams, err)
		}
	}
}

func serveHTTP(errs chan error, orgData *teams.OrgData) {
	mux := http.NewServeMux()
	mux.Handle("/teams", GetTeamHandler(orgData))
	mux.Handle("/metrics", promhttp.Handler())

	listenAt := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    listenAt,
		Handler: mux,
	}

	go func() {
		errs <- srv.ListenAndServe()
	}()
}

func doMain(cmd *cobra.Command, _ []string) error {
	errs := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	serveHTTP(errs, orgData)

	select {
	case <-stop:
		fmt.Println("Sutting down...")
		return nil
	case err := <-errs:
		fmt.Println("Failed to start server:", err.Error())
		return err
	}
}

func main() {
	cmd := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	teams.AddFlags(cmd)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Execute()
}
