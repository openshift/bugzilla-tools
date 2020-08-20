package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	//"time"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/eventlogger"

	"github.com/andygrunwald/go-jira"
	"github.com/ghodss/yaml"
	"github.com/kr/pretty"
	jiraHelper "github.com/openshift/bugzilla-tools/pkg/jira"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	port = "8002"
)

var (
	jiraQuery = fmt.Sprintf(`issuetype = Epic AND FixVersion = "OpenShift %s" AND Priority not in (Unprioritized) AND ("OpenShift Planning" != no-feature OR "OpenShift Planning" is EMPTY) AND status != "Won't Fix / Obsolete" AND filter = "Filter - Non AOS Projects"`, bugs.CurrentReleaseMinor)
)

type SnapshotData struct {
	Snapshots []string `json:"snapshots"`
}

func convert(in interface{}, out interface{}) error {
	b, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, out)
	if err != nil {
		return err
	}
	return nil
}

func getSnapshotDiff(oldDate, newDate string, cmd *cobra.Command, client *jira.Client) (*jiraHelper.JiraDiffData, error) {
	oldIssues, err := jiraHelper.GetDiskIssues(cmd, oldDate)
	if err != nil {
		return nil, err
	}

	newIssues, err := jiraHelper.GetDiskIssues(cmd, newDate)
	if err != nil {
		return nil, err
	}

	added, removed, err := jiraHelper.DiffIssueLists(oldIssues, newIssues)
	if err != nil {
		return nil, err
	}

	addedInfo, err := jiraHelper.GetIssuesInfo(client, added)
	if err != nil {
		return nil, err
	}

	removedInfo, err := jiraHelper.GetIssuesInfo(client, removed)
	if err != nil {
		return nil, err
	}

	diffData := &jiraHelper.JiraDiffData{
		NewDate: newDate,
		OldDate: oldDate,
		Added:   addedInfo,
		Removed: removedInfo,
	}
	return diffData, nil
}

func getSnapshot(cmd *cobra.Command, query url.Values, which string) (string, error) {
	snaps, ok := query[which]
	if !ok || len(snaps) != 1 || len(snaps[0]) < 1 {
		return "", fmt.Errorf("Missing %s parameter", which)
	}
	snap := snaps[0]
	// FIXME should do more validation on snap
	return snap, nil

}

func DiffHandler(cmd *cobra.Command, client *jira.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		oldDate, err := getSnapshot(cmd, query, "oldDate")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newDate, err := getSnapshot(cmd, query, "newDate")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		diffData, err := getSnapshotDiff(oldDate, newDate, cmd, client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(diffData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}
}

func GetSnapshotsHandler(cmd *cobra.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshots, err := jiraHelper.GetSnapshots(cmd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := SnapshotData{
			Snapshots: snapshots,
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func serveHTTP(errs chan error, cmd *cobra.Command, client *jira.Client) {
	mux := http.NewServeMux()
	mux.Handle("/diff", DiffHandler(cmd, client))
	mux.Handle("/snapshots", GetSnapshotsHandler(cmd))
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

type DataCollector struct {
	jiraClient *jira.Client
	cmd        *cobra.Command
}

func (dc *DataCollector) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	client := dc.jiraClient
	// Get all of the issues
	issues, err := jiraHelper.GetIssues(client, jiraQuery)
	if err != nil {
		return err
	}

	// Write the issues to disk
	err = jiraHelper.WriteIssues(dc.cmd, issues)
	if err != nil {
		return err
	}
	pretty.Println("HERE!")
	return nil
}

func CollectData(schedule []string, recorder events.Recorder, cmd *cobra.Command, jiraClient *jira.Client) factory.Controller {
	d := &DataCollector{
		cmd:        cmd,
		jiraClient: jiraClient,
	}
	return factory.New().ResyncSchedule(schedule...).WithSync(d.sync).ToController("CollectJiraData", recorder)
}

func doMain(cmd *cobra.Command, _ []string) error {
	errs := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	client, err := jiraHelper.GetClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	schedule := []string{
		//"CRON_TZ=America/New_York 0 1 * * 1-5",
		"* * * * *",
	}
	recorder := eventlogger.NewRecorder("DataCollector")
	collectData := CollectData(schedule, recorder, cmd, client)
	go collectData.Run(ctx, 1)

	serveHTTP(errs, cmd, client)
	fmt.Printf("Serving at %s\n", port)

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
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	jiraHelper.AddFlags(cmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
