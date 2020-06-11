FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bug-exportor /bug-exportor
RUN chmod +x /bug-exportor

CMD /bug-exportor --bugzilla-key=/etc/bugzilla/bugzillaKey
