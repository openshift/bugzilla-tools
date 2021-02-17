### Bugzilla Tools

Assorted tooling for monitoring and manipulating [Bugzilla](https://bugzilla.redhat.com) bugs for the OpenShift project.

### Repo Layout

The repo is a collection of tools and shared libraries. Each tool lives in cmd/* and shared code in pkg/*

Most tools have 2 Dockerfiles.
1 in cmd/*/Dockerfile
1 in Dockerfiles/Dockerfile.*

The one in cmd/*/Dockerfile is just for local testing and building.
The one in Dockerfiles/Dockerfile.* is used to build on cluster - the reason it isn't used for local building is because it does a cp of the whole repo into the build container, which is slow.

Most tools have a cmd/*/manifests/   (or cmd/*/deploment if it is old) which have the kube objects which run the tool on top of OpenShift. These are applied manually using oc apply -f. There is no automation to apply these changes.

### Adding automation to automatically run new tools

A reasonable example of adding new automation so that changes to a command are automatically applied when updated in github can be found here https://github.com/openshift/bugzilla-tools/pull/42/files

This adds an imagestream, a buildconfig, and a 'git-build-watcher'. The git-build-watcher is the magic. Since the cluster is not reachable by github we need to poll instead of get notification froma webhook. That magic polls github and forced a buildconfig to run when github has changed since the last successful build.

Most tools are deployments, not cronjobs, and thus they need a trigger to restart after a build. Pull 42 is using a cronjob so that is not present.

License
-------

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/).
