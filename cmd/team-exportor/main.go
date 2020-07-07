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

	"github.com/openshift/bugzilla-tools/pkg/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	port = "8000"
)

func GetTeamHandler(orgData *teams.OrgData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewEncoder(w).Encode(orgData)
		if err != nil {
			fmt.Errorf("Unable to encode: %v: %v", orgData, err)
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
	orgData.Reconciler()

	serveHTTP(errs, orgData)
	fmt.Println("http server started.")

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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
