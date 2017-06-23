FROM azuresdk/azure-cli-python:latest

#updateTemplate.json: is generated in the makefile and then copied to the image
#deployAzures.sh: script that logins in to azure, deploys a swarm, and runs a set of e2e tests
#swarmCountCheck.sh:  Helper script for deployAzure.sh that ensures that the entire swarm is up
#verifying the current count of worker and manager nodes

COPY tmp/updatedTemplate.json /files/swarmCountCheck.sh /files/deployAzure.sh /usr/bin/

RUN chmod 755 /usr/bin/deployAzure.sh /usr/bin/swarmCountCheck.sh


ENTRYPOINT ["/usr/bin/deployAzure.sh"]
