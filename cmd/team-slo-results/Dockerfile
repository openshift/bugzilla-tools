FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY team-slo-results /team-slo-results
RUN chmod +x /team-slo-results

COPY web/build /web/build
CMD /team-slo-results --bugzilla-key=/etc/bugzilla/bugzillaKey
