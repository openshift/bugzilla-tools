FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugs-per-team /bugs-per-team
RUN chmod +x /bugs-per-team

CMD /bugs-per-team --bugzilla-key=/etc/bugzilla/bugzillaKey
