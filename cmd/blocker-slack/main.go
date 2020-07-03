package main

import (
	"context"
	goflag "flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog"

	"github.com/eparis/bugtool/pkg/blockerslack"
	"github.com/eparis/bugtool/pkg/blockerslack/config"
	"github.com/eparis/bugtool/pkg/bugs"
	"github.com/eparis/bugtool/pkg/slack"
	"github.com/eparis/bugtool/pkg/teams"
	"github.com/eparis/bugtool/pkg/version"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := context.TODO()

	cmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "An operator that updates slack with information about bugzilla blockers",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := config.GetConfig(cmd, ctx)
			if err != nil {
				klog.Fatalf("Unable to load config: %v", err)
			}
			if err := blockerslack.Run(ctx, *c, cmd); err != nil {
				klog.Fatal(err)
			}
		},
	}

	config.AddFlags(cmd)
	slack.AddFlags(cmd)
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
