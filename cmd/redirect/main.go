package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	port   = "8000"
	oldVal = `.ocp.`
	newVal = `.ocp4.`
)

func getHost(r *http.Request) string {
	if r.URL.Host != "" {
		return r.URL.Host
	}
	return r.Host
}

func handleRedirect(rw http.ResponseWriter, req *http.Request) {
	oldHost := getHost(req)
	newHost := strings.Replace(oldHost, oldVal, newVal, -1)
	URL := req.URL.RequestURI()
	newURL := fmt.Sprintf("%s://%s%s", "http", newHost, URL)
	http.Redirect(rw, req, newURL, http.StatusSeeOther)
}

func processEnv() error {
	if p, set := os.LookupEnv("PORT"); set {
		port = p
	}
	if old, set := os.LookupEnv("OLDVAL"); set {
		oldVal = old
	}
	if new, set := os.LookupEnv("NEWVAL"); set {
		newVal = new
	}
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := processEnv(); err != nil {
				return err
			}
			http.HandleFunc("/", handleRedirect) //This is the exact url I want to redirect from
			listenURL := fmt.Sprintf(":%s", port)
			http.ListenAndServe(listenURL, nil)
			return nil
		},
	}
	cmd.Flags().StringVar(&port, "port", port, "port to listen on")
	cmd.Flags().StringVar(&oldVal, "old-val", oldVal, "pattern in old URL to replace")
	cmd.Flags().StringVar(&newVal, "new-val", newVal, "new pattern in URL to replace")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
