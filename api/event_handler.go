package main

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
)

// influx client docs: https://github.com/influxdata/influxdb-client-go#writes

type FalcoEvent struct {
	Output       string                 `json:"output"`
	Priority     string                 `json:"priority"`
	Rule         string                 `json:"rule"`
	Time         time.Time              `json:"time"`
	OutputFields map[string]interface{} `json:"output_fields"`
}

func write_event(fe FalcoEvent, oJson string, iWriter api.WriteAPI) {
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
	// new events are always false
	p.AddTag("acknowledged", "false")
	p.AddField("output", fe.Output)
	// store the original json, too
	p.AddField("json", oJson)
	iWriter.WritePoint(p)
}

func get_events(iReader api.QueryAPI, page, npp int, includeAcknowledged bool) ([]string, error) {
	log.Debug().Int("page", page).Int("npp", npp).Bool("includeAck", includeAcknowledged).Msg("starting get_events")
	offset := page * npp
	// note: we are querying where _field == json because we are going to json deserialize the events
	queryConditionals := `r._measurement == "event" and r._field == "json"`
	if !includeAcknowledged {
		queryConditionals = fmt.Sprintf(`%s and r.acknowledged == "false"`, queryConditionals)
	}
	// TODO: make the range configurable
	query := fmt.Sprintf(`from(bucket:"events") |> range(start: -1w) |> filter(fn: (r) => %s) |> limit(n: %d, offset: %d)`, queryConditionals, npp, offset)
	log.Debug().Str("query", query).Msg("running query")
	result, err := iReader.Query(context.Background(), query)
	if err != nil {
		log.Error().Err(err).Msg("error running query")
		return nil, err
	}
	res := make([]string, 0)
	for result.Next() {
		resultJson := result.Record().ValueByKey("_value")
		if resultJson != nil {
			res = append(res, fmt.Sprintf("%v", resultJson))
		}
	}
	return res, nil
}
