FROM docker4x/awscli:latest

# createStack.sh: Uses the AWS CLI to create a stack, make sure the stack comes up, and runs tests in the stack
# based on a docker template
# swarmCountCheck.sh: Helper script for createStack.sh that ensures the entire swarm is up by verifying 
# the total  count of worker and manager nodes against the expected count

RUN apk update && apk add bash openssh jq && rm -Rf /tmp/* /var/lib/cache/apk/*

COPY swarmCountCheck.sh createStack.sh /usr/bin/
RUN chmod 755 /usr/bin/createStack.sh /usr/bin/swarmCountCheck.sh


ENTRYPOINT ["/usr/bin/createStack.sh"]
