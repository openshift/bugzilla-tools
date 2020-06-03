FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugtool /bugtool
RUN chmod +x /bugtool

CMD /bugtool --test-team-data=/var/run/testTeamData.yml --bugzilla-key=/etc/bugzilla/bugzillaKey --github-key=/etc/github/githubKey
