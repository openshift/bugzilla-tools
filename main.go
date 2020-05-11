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

	"github.com/eparis/react-material/pkg/bugs"
	"github.com/eparis/react-material/pkg/metrics"
	"github.com/eparis/react-material/pkg/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func GetAPIHandler(bugData *bugs.BugData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bugMap := bugData.GetBugMap()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := json.NewEncoder(w).Encode(bugMap)
		if err != nil {
			fmt.Errorf("Unable to encode: %v: %v", bugMap, err)
		}
	}
}

func GetFileSystemHandler(directory string) http.HandlerFunc {
	dir := http.Dir(directory)
	fs := http.FileServer(dir)
	return fs.ServeHTTP
}

func serveHTTP(errs chan error, bugData *bugs.BugData) {
	port := 8000
	listenAt := fmt.Sprintf(":%d", port)

	mux := http.NewServeMux()
	path := "./react-material-ui/build"
	mux.Handle("/", GetFileSystemHandler(path))
	mux.Handle("/api", GetAPIHandler(bugData))
	mux.Handle("/metrics", promhttp.Handler())

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

	teams, err := teams.GetTeamData(cmd)
	if err != nil {
		return err
	}

	bugData := &bugs.BugData{}
	bugs.BugDataReconciler(errs, cmd, teams, bugData)

	serveHTTP(errs, bugData)

	metrics.Setup(errs, bugData, teams)

	select {
	case <-stop:
		fmt.Println("Sutting down...")
		return nil
	case err := <-errs:
		fmt.Println("Failed to start server:", err.Error())
		return err
	}
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.Execute()
}
