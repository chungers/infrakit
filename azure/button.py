#!/usr/bin/python

import urllib
import sys

if len(sys.argv) != 2:
    sys.stderr.write('Usage: python button.py [arm_template_url]\n')
    sys.stderr.write('Must pass ARM template location as argument.\n')
    exit(1)

arm_template_url = urllib.quote_plus(sys.argv[1])

if __name__ == '__main__':
    print('<a href="https://portal.azure.com/#create/Microsoft.Template/uri/{}"><img src="https://s3.amazonaws.com/docker-for-azure/deploy_to_azure.png"></a>'.format(arm_template_url))
