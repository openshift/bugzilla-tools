FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bug-automation /bug-automation
RUN chmod +x /bug-automation

COPY operations/ operations/

CMD /bug-automation --bugzilla-key=/etc/bugzilla/bugzillaKey
