package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

//GLOBALS
var influxClient influxdb2.Client
var influxWriter api.WriteAPI

func configLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
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

func setupInflux(config *Config) {
	// setup influx
	influxClient = influxdb2.NewClient(config.Influx.Url, config.Influx.Token)
	influxWriter = influxClient.WriteAPI(config.Influx.Org, config.Influx.Bucket)
}

func newEvent(w http.ResponseWriter, r *http.Request) {
	return
}

func main() {
	// configure logger and get the config
	configLogger()
	config := getConfig("./config.yml")
	// setup influx using the config values
	setupInflux(&config)
	// setup and start gorilla
	r := mux.NewRouter()
	r.HandleFunc("/event", newEvent).Methods("POST")
	serverAddr := config.Server.Address
	if config.Server.UseSSL {
		certPath := config.Server.CertPath
		keyPath := config.Server.KeyPath
		log.Info().Str("keypath", keyPath).Str("certPath", certPath).Msg("starting API server using SSL")
		err := http.ListenAndServeTLS(serverAddr, certPath, keyPath, r)
		log.Fatal().Err(err).Msg("API server stopped")
	} else {
		log.Warn().Msg("starting API server without SSL")
		err := http.ListenAndServe(serverAddr, r)
		log.Fatal().Err(err).Msg("API server stopped")
	}
}
