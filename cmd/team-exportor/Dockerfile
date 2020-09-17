FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf update -y && microdnf clean all

COPY team-exportor /team-exportor
RUN chmod +x /team-exportor

CMD /team-exportor --overwrite-team-data=/var/run/overwriteTeamData.yml --github-key=/etc/github/githubKey --data-from-github --google-sheet=/etc/google-sheet/
