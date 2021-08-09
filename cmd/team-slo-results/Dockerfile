FROM registry.access.redhat.com/ubi8/go-toolset AS builder
ENV CMD=team-slo-results
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root . ${PROJECT}
WORKDIR ${CMDDIR}
RUN make build


FROM registry.access.redhat.com/ubi8/nodejs-14 AS node-builder
ENV CMD=team-slo-results
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root ${CMDBASE}/web /web
WORKDIR /web
RUN make build


FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV CMD=team-slo-results
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

# PSI cluster is failing dnf commands right now
# RUN microdnf update -y && microdnf clean all
COPY --from=builder ${CMDDIR}/${CMD} /${CMD}
RUN chmod +x /${CMD}

COPY --from=node-builder /web/build /web/build
CMD /${CMD} --bugzilla-key=/etc/bugzilla/bugzillaKey
