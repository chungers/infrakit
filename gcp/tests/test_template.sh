#!/bin/bash

set -ex

python files/apply.py Docker.jinja > template.yaml

diff -b template.yaml files/expected.template.yaml
