FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugtool /bugtool
RUN chmod +x /bugtool

COPY operations/ operations/

CMD /bugtool
