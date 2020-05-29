FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugautomation /bugautomation
RUN chmod +x /bugautomation

COPY operations/ operations/

CMD /bugautomation
