FROM registry.access.redhat.com/ubi8/go-toolset AS builder
ENV CMD=bug-automation
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root . ${PROJECT}
WORKDIR ${CMDDIR}
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV CMD=bug-automation
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

# PSI cluster is failing dnf commands right now
# RUN microdnf update -y && microdnf clean all
COPY --from=builder ${CMDDIR}/${CMD} /${CMD}
RUN chmod +x /${CMD}
COPY --from=builder ${CMDDIR}/operations/ operations/
CMD /${CMD} --bugzilla-key=/etc/bugzilla/bugzillaKey
