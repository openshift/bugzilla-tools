package config

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/openshift/library-go/pkg/controller/fileobserver"
	"github.com/spf13/cobra"
	"k8s.io/klog"
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

// GetConfig will populate `cfg` with the contents in the file specified by the `flagname` in from `cmd`
// It will also start watching the file and will exit the program if the file changes
func GetConfig(cmd *cobra.Command, flagName string, ctx context.Context, cfg interface{}) error {
	configPath, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return err
	}

	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	go restartOnConfigChange(ctx, configPath, configBytes)

	return yaml.Unmarshal(configBytes, cfg)
}

func Decode(s string) string {
	if strings.HasPrefix(s, "base64:") {
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, "base64:"))
		if err != nil {
			return s
		}
		return string(data)
	}
	return s
}
