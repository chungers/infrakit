#!/bin/sh

./bake-ami.sh "$@" 1>&2
if [ "$1" = "bake" ]
then
	if [ $LOAD_IMAGES == 'true' ]; then
		cat /build/ami_id_ee.out
	else
		cat /build/ami_id.out
	fi
fi
