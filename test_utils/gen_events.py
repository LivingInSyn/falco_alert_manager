'''
curl -XPOST -d '{"output":"16:31:56.746609046: Error File below a known binary directory opened for writing (user=root command=touch /bin/hack file=/bin/hack)","priority":"Error","rule":"Write below binary dir","time":"2022-06-26T23:31:56.746609046Z", "output_fields": {"evt.time":1507591916746609046,"fd.name":"/bin/hack","proc.cmdline":"touch /bin/hack","user.name":"root"}}' -H "Content-Type: application/json" http://localhost:8081/event
'''

import requests
from datetime import datetime,timedelta
import time
import json
import sys

def gen_event(num_events=50):
    # 2022-06-26T23:31:56.746609046Z
    # 2022-06-27T21:15:49.041680000Z
    # 1656379146166481000
    # 1656393732429092864
    # 1507591916746609046
    event = {
        "output": "{time}: Error File below a known binary directory opened for writing (user=root command=touch /bin/hack file=/bin/hack)",
        "priority": "Error",
        "rule": "Write below binary dir",
        "time": "",
        "output_fields": {
            "evt.time": 1507591916746609046,
            "fd.name": "/bin/hack",
            "proc.cmdline": "touch /bin/hack",
            "user.name": "root"
        }
    }
    stime = datetime.now() - timedelta(hours=num_events+1)
    for _ in range(0, num_events):
        # create a copy and update the fields
        sevent = event.copy()
        output_time = stime.strftime('%H:%M:%S.%f000')
        time_field = f'{stime.isoformat()}000Z'
        unix_time = int(datetime.utcnow().timestamp() * 1000000000)
        sevent['output'] = sevent['output'].replace('{time}', output_time)
        sevent['time'] = time_field
        sevent['output_fields']['evt.time'] = unix_time
        # send it
        requests.post('http://localhost:8081/event', data=json.dumps(sevent))
        # update stime
        stime = stime + timedelta(hours=1)
        time.sleep(.2)

if len(sys.argv) > 1:
    num_events = sys.argv[1]
    num_events = int(num_events)
else:
    num_events = 50
gen_event(num_events)
