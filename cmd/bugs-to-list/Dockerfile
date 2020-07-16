FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugs-to-list /bugs-to-list
RUN chmod +x /bugs-to-list

CMD /bugs-to-list --bugzilla-key=/etc/bugzilla/bugzillaKey
