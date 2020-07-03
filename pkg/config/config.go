package config

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"strings"
	"time"

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

// GetBytes will return the bytes from the file specified in a flag named flagName.
// It will also start a watch on the file which will terminate the program if the
// file changes.
func GetBytes(cmd *cobra.Command, flagName string, ctx context.Context) ([]byte, error) {
	configPath, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return nil, err
	}

	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	go restartOnConfigChange(ctx, configPath, configBytes)
	return configBytes, nil
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
