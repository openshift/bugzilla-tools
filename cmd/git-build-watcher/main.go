package main

//This was shameless (or shamefully, depending on how you look at it) stolen in entirety from
//https://raw.githubusercontent.com/dmage/git-build-watcher/

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	buildv1 "github.com/openshift/api/build/v1"
	buildclient "github.com/openshift/client-go/build/clientset/versioned"
)

func getCommitish(repo, ref string) (string, error) {
	klog.V(6).Infof("Executing: git ls-remote %s %s", repo, ref)
	cmd := exec.Command("git", "ls-remote", repo, ref)
	cmd.Stderr = os.Stderr
	buf, err := cmd.Output()
	if err != nil {
		return "", err
	}
	output := string(buf)
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 1 {
		return "", fmt.Errorf("unexpected output from git ls-remote: %s", output)
	}
	idx := strings.Index(lines[0], "\t")
	if idx == -1 {
		return "", fmt.Errorf("unexpected output from git ls-remote: %s", output)
	}
	commitish := lines[0][:idx]
	klog.V(6).Infof("Got commitish for %s %s: %s", repo, ref, commitish)
	return commitish, nil
}

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <buildconfig>\n", os.Args[0])
		pflag.PrintDefaults()
	}

	klog.InitFlags(nil)
	configFlags := genericclioptions.NewConfigFlags(false)
	configFlags.AddFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	args := pflag.Args()
	if len(args) != 1 {
		klog.Info(args)
		pflag.Usage()
		os.Exit(1)
	}
	buildConfigName := args[0]

	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		klog.Exit(err)
	}

	namespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		klog.Exit(err)
	}

	ctx := context.Background()
	buildClient := buildclient.NewForConfigOrDie(restConfig)

	buildConfig, err := buildClient.BuildV1().BuildConfigs(namespace).Get(ctx, buildConfigName, metav1.GetOptions{})
	if err != nil {
		klog.Exit(err)
	}

	if buildConfig.Spec.Source.Type != buildv1.BuildSourceGit {
		klog.Exitf("%s uses %s source, but this tool works only with Git sources", buildConfigName, buildConfig.Spec.Source.Type)
	}

	if buildConfig.Spec.Source.Git == nil {
		klog.Exitf("error: spec.source is %s, but spec.source.git is not set", buildConfig.Spec.Source.Type)
	}

	commitish, err := getCommitish(buildConfig.Spec.Source.Git.URI, buildConfig.Spec.Source.Git.Ref)
	if err != nil {
		klog.Exit(err)
	}

	builds, err := buildClient.BuildV1().Builds(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "openshift.io/build-config.name=" + buildConfigName,
	})
	if err != nil {
		klog.Exit(err)
	}

	klog.V(4).Infof("Found %d builds", len(builds.Items))

	if len(builds.Items) > 0 {
		latest := &builds.Items[0]
		for i, build := range builds.Items {
			if build.CreationTimestamp.After(latest.CreationTimestamp.Time) {
				latest = &builds.Items[i]
			}
		}

		klog.V(4).Infof("Build %s is identified as the latest build", latest.Name)

		latestCommit := ""
		if latest.Spec.Revision == nil || latest.Spec.Revision.Git == nil {
			if latest.Status.Phase == buildv1.BuildPhaseNew ||
				latest.Status.Phase == buildv1.BuildPhasePending ||
				latest.Status.Phase == buildv1.BuildPhaseRunning {
				klog.V(2).Infof("The build %s is %s and does not have commit information. Please re-run this tool when its status changes.", latest.Name, latest.Status.Phase)
				os.Exit(0)
			}
			klog.V(2).Infof("The latest build %s does not have revision commit", latest.Name)
		} else {
			latestCommit = latest.Spec.Revision.Git.Commit
		}

		if latestCommit == commitish {
			klog.V(2).Infof("The commitish %s has already been built. Nothing to do.", commitish)
			os.Exit(0)
		}
	}

	klog.V(2).Infof("Triggering a new build for the commitish %s", commitish)

	build, err := buildClient.BuildV1().BuildConfigs(namespace).Instantiate(ctx, buildConfigName, &buildv1.BuildRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: buildConfigName,
		},
		TriggeredBy: []buildv1.BuildTriggerCause{
			{
				Message: fmt.Sprintf("git-build-watcher detected new commit %s", commitish),
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		klog.Exit(err)
	}

	fmt.Printf("build/%s\n", build.Name)
}
