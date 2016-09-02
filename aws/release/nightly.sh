#!/bin/bash
echo "================================================"
echo ""
echo ""
echo "Start nightly build"
echo $PATH
PATH=$PATH:/usr/local/bin
echo $PATH

cd /home/ubuntu/code/moby/alpine
make ami

echo "Finished AMI build, lets move onto next part."

# get ami-id
# look for anyfiles that look like ami_*.out those are ami's ready to be processed.
for f in /home/ubuntu/out/ami_*.out; do

    ## Check if the glob gets expanded to existing files.
    ## If not, f here will be exactly the pattern above
    ## and the exists test will evaluate to false.
    if [ -e "$f" ]; then
        AMI_ID=$(cat $f)
        mv $f $f.done
    fi

    ## We can only process one at a time.
    break
done

if [[ -z $AMI_ID ]]
then
    echo "There is no AMI_ID, nothing to do, so stopping."
    # there are no AMI's ready skip.
     exit 1
fi
echo $AMI_ID

AMI_SOURCE_REGION=us-east-1
DOCKER_AWS_ACCOUNT_URL=https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt

export AWS_ACCESS_KEY_ID=$(cat ~/.aws/credentials | grep aws_access_key_id | cut -f2 -d= | sed -e 's/^[ \t]*//')
export AWS_SECRET_ACCESS_KEY=$(cat ~/.aws/credentials | grep aws_secret_access_key | cut -f2 -d= | sed -e 's/^[ \t]*//')

# git update
cd /home/ubuntu/code/editions/
git pull

# get docker version

cd /tmp
curl -s -o docker https://master.dockerproject.org/linux/amd64/docker
chmod 755 docker
export DOCKER_VERSION=$(./docker version -f '{{.Client.Version}}')
rm -f docker

NOW=$(date +"%m_%d_%Y")
export EDITION_VERSION=nightly_$NOW

export VERSION=aws-v$DOCKER_VERSION-$EDITION_VERSION

# build binaries
cd /home/ubuntu/code/editions/tools/buoy/
./build_buoy.sh

cd /home/ubuntu/code/editions/aws/dockerfiles/

# build images

./build_and_push_all.sh

cd /home/ubuntu/code/editions/aws/release

# run release
./run_release.sh -d $DOCKER_VERSION -e $EDITION_VERSION -a $AMI_ID -r $AMI_SOURCE_REGION -c nightly -l $DOCKER_AWS_ACCOUNT_URL -u nightly

# run cleanup, remove things that are more than X days old.
python cleanup.py

# run tests
python test_cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/nightly/latest.json

# Rebuild the nightly index page.
python build_index.py

# notify results
python notify.py
