package main

import (
	"errors"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// influx client docs: https://github.com/influxdata/influxdb-client-go#writes

type FalcoEvent struct {
	Output       string                 `json:"output"`
	Priority     string                 `json:"priority"`
	Rule         string                 `json:"rule"`
	Time         time.Time              `json:"time"`
	OutputFields map[string]interface{} `json:"output_fields"`
}

func write_event(fe FalcoEvent, iWriter api.WriteAPI) {
	// sample event: https://falco.org/docs/alerts/#program-output-sending-alerts-to-network-channel
	/*
		{
			"output" : "16:31:56.746609046: Error File below a known binary directory opened for writing (user=root command=touch /bin/hack file=/bin/hack)"
			"priority" : "Error",
			"rule" : "Write below binary dir",
			"time" : "2017-10-09T23:31:56.746609046Z",
			"output_fields" : {
				"user.name" : "root",
				"evt.time" : 1507591916746609046,
				"fd.name" : "/bin/hack",
				"proc.cmdline" : "touch /bin/hack"
			}
		}
	*/
	p := influxdb2.NewPoint("event",
		map[string]string{"priority": fe.Priority, "rule": fe.Rule},
		fe.OutputFields,
		fe.Time)
	p.AddTag("acknowledged", "false")
	p.AddField("output", fe.Output)
	iWriter.WritePoint(p)
}

func get_events(iReader api.QueryAPI, page, npp int, includeAcknowledged bool) ([]FalcoEvent, error) {
	return nil, errors.New("not implemented")
}
