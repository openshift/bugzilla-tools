FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY upcoming-sprint-stats /upcoming-sprint-stats
RUN chmod +x /upcoming-sprint-stats

CMD /upcoming-sprint-stats --bugzilla-key=/etc/bugzilla/bugzillaKey
