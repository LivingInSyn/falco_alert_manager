package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

//GLOBALS
var ctx context.Context
var timescaleConn *pgxpool.Pool

// setup max, min int sizes
const UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64
const (
	MaxInt  = 1<<(UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
	MinInt  = -MaxInt - 1         // -1 << 31 or -1 << 63
	MaxUint = 1<<UintSize - 1     // 1<<32 - 1 or 1<<64 - 1
)

func configLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	_, dob := os.LookupEnv("FAM_DEBUG")
	if dob {
		log.Info().Msg("Log level set to DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Info().Msg("Log level set to default")
	}
	log.Info().Msg("Logger setup")
}

func getConfig(configPath string) Config {
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't open config file")
	}
	defer f.Close()
	var config Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Coudln't unmarshal config file")
	}
	return config
}

func setupTimescale(config *Config) {
	ctx = context.Background()
	connStr := fmt.Sprintf("postgres://%s:%s@%s/fam", config.Timescale.Username, config.Timescale.Password, config.Timescale.Url)
	var err error
	timescaleConn, err = pgxpool.Connect(ctx, connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func getQueryInt(val string, dval, min, max int) (int, error) {
	if val == "" {
		return dval, nil
	}
	ival, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	if ival < max && ival >= min {
		return ival, nil
	}
	if ival >= min && ival > max {
		return max, nil
	}
	if ival < min && ival < max {
		return min, nil
	}
	return -1, errors.New("range error")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func newEvent(w http.ResponseWriter, r *http.Request) {
	// get the event from the body and parse it
	var fe FalcoEvent
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	bodyText := string(bodyBytes)
	err = json.Unmarshal(bodyBytes, &fe)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	log.Debug().Interface("event", fe).Msg("got new falco event")
	go write_event(fe, bodyText, timescaleConn, ctx)
}

func paginatedEvent(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("starting paginatedEvent")
	page := r.URL.Query().Get("page")
	pageVal, err := getQueryInt(page, 0, -1, MaxInt)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid page number")
		return
	}
	npp := r.URL.Query().Get("per")
	numPerPage, err := getQueryInt(npp, 25, 1, 50)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid page number")
		return
	}
	ia := r.URL.Query().Get("includeAcknowledged")
	if ia == "" {
		ia = "false"
	}
	includeAcknowledged, err := strconv.ParseBool(ia)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid argument")
		return
	}
	events, err := get_events(pageVal, numPerPage, includeAcknowledged, timescaleConn, ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get events)")
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	respondWithJSON(w, 200, events)
}

func main() {
	// configure logger and get the config
	configLogger()
	config := getConfig("./config.yml")
	// setup influx using the config values
	setupTimescale(&config)
	// setup and start gorilla
	r := mux.NewRouter()
	r.HandleFunc("/event", newEvent).Methods("POST")
	log.Debug().Msg("setting up event paged")
	r.HandleFunc("/event_paged", paginatedEvent).Methods("GET")
	serverAddr := config.Server.Address
	if config.Server.UseSSL {
		certPath := config.Server.CertPath
		keyPath := config.Server.KeyPath
		log.Info().Str("address", serverAddr).Str("keypath", keyPath).Str("certPath", certPath).Msg("starting API server using SSL")
		err := http.ListenAndServeTLS(serverAddr, certPath, keyPath, r)
		log.Fatal().Err(err).Msg("API server stopped")
	} else {
		log.Warn().Str("address", serverAddr).Msg("starting API server without SSL")
		err := http.ListenAndServe(serverAddr, r)
		log.Fatal().Err(err).Msg("API server stopped")
	}
}
