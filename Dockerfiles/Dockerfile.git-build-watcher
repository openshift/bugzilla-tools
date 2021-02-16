FROM registry.access.redhat.com/ubi8/go-toolset AS builder
ENV CMD=git-build-watcher
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root . ${PROJECT}
WORKDIR ${CMDDIR}
RUN make build


FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV CMD=git-build-watcher
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

RUN microdnf update -y && microdnf install -y git && microdnf clean all
COPY --from=builder ${CMDDIR}/${CMD} /${CMD}
RUN chmod +x /${CMD}
CMD /${CMD}
