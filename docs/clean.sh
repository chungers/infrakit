#!/bin/bash
echo cleaning $1
rm -f $1/*.old
rm -f $1/dist.zip
rm -f $1/dist/*.html
rm -rf $1/src
rm -f $1/Gruntfile.js
rm -f $1/bower.json
