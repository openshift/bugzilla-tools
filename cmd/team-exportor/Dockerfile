FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY team-exportor /team-exportor
RUN chmod +x /team-exportor

CMD /team-exportor --test-team-data=/var/run/overwriteTeamData.yml --github-key=/etc/github/githubKey
