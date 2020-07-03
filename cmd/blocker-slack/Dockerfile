FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY blocker-slack /blocker-slack
RUN chmod +x /blocker-slack

CMD /blocker-slack --slack-key=/etc/slack/slackKey 
