FROM registry.access.redhat.com/ubi8/go-toolset AS builder
ENV CMD=redirect
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root . ${PROJECT}
WORKDIR ${CMDDIR}
RUN make build


FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV CMD=redirect
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

RUN microdnf update -y && microdnf clean all
COPY --from=builder ${CMDDIR}/${CMD} /${CMD}
RUN chmod +x /${CMD}
CMD /${CMD}
