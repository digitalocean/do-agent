#!/usr/bin/env python
# vim: set ts=4 sts=4 sw=4 expandtab :

'''
uat.sh dynamic inventory script for Ansible, in Python.
'''

import os
import sys
import argparse
import subprocess
import json
import urllib2

try:
    import json
except ImportError:
    import simplejson as json

class ExampleInventory(object):

    def __init__(self):
        self.inventory = {}
        self.read_cli_args()

        if bool(self.args.list) == bool(self.args.host):
            sys.exit("script must be executed with either '--list' or '--host <name>'")
        elif self.args.list:
            # Called with `--list`.
            self.inventory = self.remote_inventory()
        elif self.args.host:
            # Called with `--host [hostname]`.
            # Not implemented, since we return _meta info `--list`.
            self.inventory = self.empty_inventory()
        else:
            # If no groups or vars are present, return an empty inventory.
            self.inventory = self.empty_inventory()

        print json.dumps(self.inventory);

    # Example inventory for testing.
    def remote_inventory(self):
        headers = {'Accept': 'application/json',
                'Authorization': 'Bearer '+os.getenv('AUTH_TOKEN', '')}
        req = urllib2.Request('https://api.digitalocean.com/v2/droplets?tag_name=do-agent-uat-{tag}'.format(tag=os.getenv('USERNAME', 'nobody')), None, headers)
        resp = urllib2.urlopen(req)
        body = json.loads(resp.read())
        droplets = body['droplets']
        if len(droplets) < 1:
            return self.empty_inventory()
        sys.exit(json.dumps(droplets[0], indent=4, sort_keys=True))
        return {
            'systemd': {
                'hosts': ['192.168.28.71', '192.168.28.72'],
                'vars': {
                    'ansible_ssh_user': 'root',
                }
            },
            'upstart': {
                'hosts': ['192.168.28.71', '192.168.28.72'],
                'vars': {
                    'ansible_ssh_user': 'root',
                }
            },
            'deb': {
                'hosts': ['192.168.28.71', '192.168.28.72'],
                'vars': {
                    'ansible_ssh_user': 'root',
                }
            },
            'rpm': {
                'hosts': ['192.168.28.71', '192.168.28.72'],
                'vars': {
                    'ansible_ssh_user': 'root',
                }
            },
            '_meta': {
                'hostvars': {
                    '192.168.28.71': {
                        'host_specific_var': 'foo'
                    },
                    '192.168.28.72': {
                        'host_specific_var': 'bar'
                    }
                }
            }
        }

    # Empty inventory for testing.
    def empty_inventory(self):
        return {'_meta': {'hostvars': {}}}

    # Read the command line args passed to the script.
    def read_cli_args(self):
        parser = argparse.ArgumentParser()
        parser.add_argument('--list', action = 'store_true')
        parser.add_argument('--host', action = 'store')
        self.args = parser.parse_args()

# Get the inventory.
ExampleInventory()
