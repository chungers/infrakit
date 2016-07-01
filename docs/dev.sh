#!/bin/bash
$@ &
while :
do
    inotifywait -r /docsdev
    rsync -av /docsdev/ /docs
done
