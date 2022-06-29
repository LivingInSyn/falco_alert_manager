package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

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

type FalcoEvent struct {
	Output       string                 `json:"output"`
	Priority     string                 `json:"priority"`
	Rule         string                 `json:"rule"`
	Time         time.Time              `json:"time"`
	OutputFields map[string]interface{} `json:"output_fields"`
}
type EventQueryResult struct {
	Time      time.Time
	Priority  string
	Rule      string
	Output    string
	FullEvent string
	Ack       bool
}

func create_table(p *pgxpool.Pool, ctx context.Context) {
	createTable := `CREATE TABLE IF NOT EXISTS event (
		time TIMESTAMPTZ NOT NULL,
		priority TEXT NOT NULL,
		rule TEXT NOT NULL,
		output TEXT NOT NULL,
		eventj json,
		ack BOOLEAN DEFAULT FALSE
	);
	SELECT create_hypertable('event', 'time', if_not_exists => TRUE);`
	_, err := p.Exec(ctx, createTable)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create timescaledb table")
	}
}

func write_event(fe FalcoEvent, oJson string, p *pgxpool.Pool, ctx context.Context) {
	insertStmt := `INSERT INTO event (time, priority, rule, output, eventj) VALUES ($1,$2,$3,$4,$5);`
	_, err := p.Exec(ctx, insertStmt, fe.Time, fe.Priority, fe.Rule, fe.Output, oJson)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert into database")
	}
}

func get_events(page, npp int, includeAcknowledged bool, p *pgxpool.Pool, ctx context.Context) ([]FalcoEvent, error) {
	log.Debug().Int("page", page).Int("npp", npp).Bool("includeAck", includeAcknowledged).Msg("starting get_events")
	offset := page * npp
	query := `SELECT time,priority,rule,output,eventj,ack FROM event WHERE`
	if !includeAcknowledged {
		query = fmt.Sprintf("%s ack='false'", query)
	} else {
		query = fmt.Sprintf("%s ack='true'", query)
	}
	// npp and offset are safe from injection since they're verified ints here
	query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, npp, offset)
	rows, err := p.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("error running paginated query")
		return nil, err
	}
	defer rows.Close()
	var results []EventQueryResult
	for rows.Next() {
		//TODO: make a type and then setup the values from the row
		var r EventQueryResult
		err = rows.Scan(&r.Time, &r.Priority, &r.Rule, &r.Output, &r.FullEvent, &r.Ack)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse row")
		}
		results = append(results, r)
	}
	var robj []FalcoEvent
	for _, r := range results {
		var fe FalcoEvent
		if err := json.Unmarshal([]byte(r.FullEvent), &fe); err != nil {
			log.Error().Str("input", r.FullEvent).Msg("failed to deserialize event from DB")
		} else {
			robj = append(robj, fe)
		}
	}
	return robj, nil
}
