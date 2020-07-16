FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugs-to-teams /bugs-to-teams
RUN chmod +x /bugs-to-teams

CMD /bugs-to-teams --bugzilla-key=/etc/bugzilla/bugzillaKey
