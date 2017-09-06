#!/bin/sh

echo "Checking if we previously built: ${PROVIDER}/${COMMIT}/${OUTPUT}"
METADATA=$(docker run --rm docker4x/awscli:latest s3api --no-sign-request get-object --bucket docker-ci-editions --key ${PROVIDER}/${COMMIT}/${OUTPUT} docker.out)
if [ $? -ne 0 ];
then
	# File doesn't exist - It's never been built with this editions commit - go ahead
	exit 0
fi
META_MOBY_COMMIT=$(echo $METADATA | jq -r '.Metadata.moby_commit')
META_DOCKER_VERSION=$(echo $METADATA | jq -r '.Metadata.docker_version')

# Check moby commit in metadata is the same
if [ "$MOBY_COMMIT" = "$META_MOBY_COMMIT" ];
then
	# It's already built - don't build again
	exit 1
fi

# It's never been built with this moby commit - go ahead
exit 0
