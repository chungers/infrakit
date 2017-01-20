#!/bin/bash

set -ex

python files/apply.py swarm.jinja > template.yaml

diff template.yaml files/expected.template.yaml
