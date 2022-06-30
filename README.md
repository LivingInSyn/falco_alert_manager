# Falco Alert Manager
Simply manage alerts from falco

## API methods currently implemented
### GET /event
A paginated list of events. Accepts the following args:

* page - which page of results (default: 0)
* per - number of events per page (max: 50, Min: 1, default: 25)
* includeAcknowledged - include or exclude acknowledged events. (Defaults `false`)

Sample:
```sh
curl "localhost:8081/event?page=0&per=5&includeAcknowledged=false"
```

Returns an object in the form:
```json
{
    "ID": "guid",
    "event": {
        <<falco event>>
    },
    "ack": true,
    "comment": "some comment"
}
```

### POST /event
Put a new event into the system. Expects a POST with a Falco event in JSON format. 

Sample:
```sh
curl -XPOST -d '{"output":"16:31:56.746609046: Error File below a known binary directory opened for writing (user=root command=touch /bin/hack file=/bin/hack)","priority":"Error","rule":"Write below binary dir","time":"2022-06-26T23:31:56.746609046Z", "output_fields": {"evt.time":1507591916746609046,"fd.name":"/bin/hack","proc.cmdline":"touch /bin/hack","user.name":"root"}}' -H "Content-Type: application/json" http://localhost:8081/event
```

### PUT /event/ack/{eventID}
Acknowledge an event

Requires a json object in the form:
```json
{
    "comment": "some comment"
}
```

Sample:
```sh
curl -s -XPUT -H "Content-Type: application/json" -d '{"comment":"foo bar baz"}' "http://localhost:8081/event/ack/05e33201-f8db-4ae5-b838-268965f11c59"
```