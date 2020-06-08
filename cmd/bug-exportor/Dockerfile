FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY bugexportor /bugexportor
RUN chmod +x /bugexportor

CMD /bugexportor --test-team-data=/var/run/testTeamData.yml --bugzilla-key=/etc/bugzilla/bugzillaKey --github-key=/etc/github/githubKey
