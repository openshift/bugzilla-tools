FROM registry.access.redhat.com/ubi8/go-toolset AS builder
ENV CMD=blocker-slack
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

COPY --chown=default:root . ${PROJECT}
WORKDIR ${CMDDIR}
RUN make build

FROM registry.access.redhat.com/ubi8/ubi:latest
ENV CMD=blocker-slack
ENV GOPATH=/go
ENV PROJECT=${GOPATH}/src/github.com/openshift/bugzilla-tools/
ENV CMDBASE=cmd/${CMD}/
ENV CMDDIR=${PROJECT}/${CMDBASE}/

# PSI cluster is failing dnf commands right now
# RUN microdnf update -y && rpm -e --justdb --nodeps tzdata && microdnf install -y tzdata && microdnf clean all
COPY --from=builder ${CMDDIR}/${CMD} /${CMD}
RUN chmod +x /${CMD}
CMD /${CMD} --bugzilla-key=/etc/bugzilla/bugzillaKey --slack-key=/etc/slack/slackKey --config=/etc/blocker-slack/config.yaml
