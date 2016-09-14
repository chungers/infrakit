#!/bin/sh

# Script to pull diagnostic info for sessions (possibly across multiple hosts)
# from S3 and push it to Docker image tags for easier inspection.
#
# Usage: ./bundle-diagnostics.sh
#
# To get detailed (redundant) logs, specify environment variable DEBUG=1

set -e

# Defines how long to wait before processing files for a session (allows time
# for all hosts in a session to come in)
LIMIT=${LIMIT:-"-5 minute"}
LIMITFILE="/tmp/limit"

# Where to store the bucket files while processing
BUCKET="/tmp/bucket"

# Repository to push resulting Docker image "bundle" to
IMAGE="docker4x/diagnostics"

arrowecho()
{
	echo " --->" "$@"
}

debugecho()
{
	if [ ! -z "${DEBUG}" ]
	then
		arrowecho "$@"
	fi
}

process()
{
	SESSION_ID="$1"
	REPORT="$2"
	TMPDIR="/tmp/${SESSION_ID}"
	DF="${TMPDIR}/Dockerfile"
	if [ ! -d "${TMPDIR}" ]
	then
		mkdir "${TMPDIR}"
	fi
	echo "FROM alpine" >"${DF}"
	for HOST in ${BUCKET}/${SESSION_ID}*
	do
		HOSTNAME=$(basename "${HOST}" | sed 's/.tar//g' | sed "s/${SESSION_ID}-//g")
		debugecho "cp \"${HOST}\" \"${TMPDIR}/${HOSTNAME}\""
		cp "${HOST}" "${TMPDIR}/${HOSTNAME}"

		# Note the use of ADD here.  This untars the archive
		# automatically.
		echo "ADD ${HOSTNAME} /info/${HOSTNAME}" >>"${DF}"
	done
	echo "WORKDIR /info" >>"${DF}"
	echo "CMD [\"sh\"]" >>"${DF}"
	arrowecho "Generated Dockerfile:"
	cat "${DF}" | sed 's/^/   /'
	echo
	debugecho "docker build -t \"${IMAGE}:${SESSION_ID}\" \"${DF}\" \"${TMPDIR}\""
	docker build -t "${IMAGE}:${SESSION_ID}" -f "${DF}" "${TMPDIR}"
	IMAGETAG="${IMAGE}:${SESSION_ID}"
	docker push "${IMAGETAG}"
	docker rmi "${IMAGETAG}"
	for HOST in ${BUCKET}/${SESSION_ID}*
	do
		BUCKETKEY=$(basename "${HOST}")
		rm "${HOST}"
		aws s3 rm "s3://editionsdiagnostics/${BUCKETKEY}"
	done
	rm -r "${TMPDIR}"
}

if [ ! -S /var/run/docker.sock ]
then
	>&2 echo "No Docker socket!"
	exit 1
fi

if [ ! -d "${BUCKET}" ]
then
	>&2 echo "No ${BUCKET} directory"
	exit 1
fi

while true
do
	debugecho "Pulling from S3 bucket..."
	S3_PULL=$(aws s3 sync --delete s3://editionsdiagnostics "${BUCKET}")
	if [ $? -ne 0 ]
	then
	>&2 echo "Pulling from S3 failed.  Did you set valid AWS credentials?"
	exit 1
	fi

	for REPORT in ${BUCKET}/*.tar
	do
		if [ "${REPORT}" = "${BUCKET}/*.tar" ]
		then
			debugecho "No files found in ${BUCKET}. Nothing to do."
			sleep 5
			continue
		fi

		REPORT_BASE=$(basename "${REPORT}")

		# Get session ID (files are named session-hostname.tar)
		SESSION_ID="$(echo "${REPORT_BASE}" | cut -d'-' -f1)"

		# Inspired by
		# http://stackoverflow.com/questions/205666/what-is-the-best-way-to-perform-timestamp-comparison-in-bash
		#
		# Thanks Bruno De Fraine.
		touch -d "${LIMIT}" "${LIMITFILE}"

		# Only process these files if they've been around a while
		# (i.e., if they're "older than" a given time limit).
		#
		# Otherwise, we might make a bundle for a session before
		# allowing all of the associated hosts adequate time to "check
		# in".
		if [ "${LIMITFILE}" -nt "${REPORT}" ]
		then
			debugecho "process \"${SESSION_ID}\" \"${REPORT_BASE}\""
			process "${SESSION_ID}" "${REPORT_BASE}"
			debugecho "Finished processing diagnostics bundle for image tag ${SESSION_ID}"
		else
			debugecho "Skipping ${REPORT} since it hasn't been long enough yet"
		fi
	done

	# Avoid AWS rate limit.
	sleep 5
done
