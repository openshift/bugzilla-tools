FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY smartsheets /smartsheets
RUN chmod +x /smartsheets

CMD /smartsheets --bugzilla-key=/etc/bugzilla/bugzillaKey --smartsheet-key=/etc/smartsheets/smartsheetKey
