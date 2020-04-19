#!/usr/bin/python

from ansible.module_utils.basic import *
import requests
import os
import stat

def make_executable(path):
    mode = os.stat(path).st_mode
    mode |= (mode & 0o444) >> 2    # copy R bits to X
    os.chmod(path, mode)

def main():
    fields = {
        "rules_path": {"default": True, "type": "str"},
    }

    module = AnsibleModule(argument_spec=fields)
    execPath = "/tmp/prom_merge"
    url = "https://github.com/lablabs/prometheus-alert-overrider/releases/download/v0.2.0/prometheus_merger"

    r = requests.get(url, allow_redirects=True)
    open(execPath, 'wb').write(r.content)

    make_executable(execPath)

    args = (execPath, module.params["rules_path"])

    popen = subprocess.Popen(args, stdout=subprocess.PIPE)
    popen.wait()
    output = popen.stdout.read()
    # output = popen.stderr.read()
    module.exit_json(changed=True, alerts=output)

if __name__ == '__main__':
    main()