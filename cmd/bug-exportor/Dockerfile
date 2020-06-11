FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bug-exportor /bug-exportor
RUN chmod +x /bug-exportor

CMD /bug-exportor --test-team-data=/var/run/testTeamData.yml --bugzilla-key=/etc/bugzilla/bugzillaKey --github-key=/etc/github/githubKey
