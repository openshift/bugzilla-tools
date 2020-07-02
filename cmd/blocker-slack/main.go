package main

import (
	"context"
	goflag "flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/library-go/pkg/controller/fileobserver"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog"

	"github.com/eparis/bugtool/pkg/blockerslack"
	"github.com/eparis/bugtool/pkg/blockerslack/config"
	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/teams"
	"github.com/eparis/bugtool/pkg/version"
)

func restartOnConfigChange(ctx context.Context, path string, startingContent []byte) {
	observer, err := fileobserver.NewObserver(1 * time.Second)
	if err != nil {
		panic(err)
	}
	if len(startingContent) == 0 {
		klog.Warningf("No configuration file available")
	}
	observer.AddReactor(func(file string, action fileobserver.ActionType) error {
		os.Exit(0)
		return nil
	}, map[string][]byte{
		path: startingContent,
	}, path)
	observer.Run(ctx.Done())
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := context.TODO()

	var configPath string

	cmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "An operator that updates slack with information about bugzilla blockers",
		Run: func(cmd *cobra.Command, args []string) {
			configBytes, _ := ioutil.ReadFile(configPath)
			go restartOnConfigChange(ctx, configPath, configBytes)
			c := &config.OperatorConfig{}
			if err := yaml.Unmarshal(configBytes, c); err != nil {
				klog.Fatalf("Unable to parse config: %v", err)
			}
			if err := blockerslack.Run(ctx, *c, cmd); err != nil {
				klog.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to operator config")
	bugs.AddFlags(cmd)
	teams.AddFlags(cmd)

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
