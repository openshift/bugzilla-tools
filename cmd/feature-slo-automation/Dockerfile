FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && rpm -e --justdb --nodeps tzdata && microdnf install -y tzdata && microdnf clean all

COPY blocker-slack /blocker-slack
RUN chmod +x /blocker-slack

CMD /blocker-slack --bugzilla-key=/etc/bugzilla/bugzillaKey --slack-key=/etc/slack/slackKey --config=/etc/blocker-slack/config.yaml
