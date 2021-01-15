FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV CMD=jira-daily-diff

RUN microdnf update -y && microdnf clean all
COPY ./${CMD} /${CMD}
RUN chmod +x /${CMD}

COPY ./web/build /web/build
CMD /${CMD} --jira-key=/etc/jira/jiraKey
