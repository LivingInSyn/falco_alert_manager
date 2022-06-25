package main

import (
	"net/http"

	"github.com/gorilla/mux"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//GLOBALS
var influxClient influxdb2.Client
var influxWriter api.WriteAPI

func configLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Str("logger setup")
}

func getConfig() {
	viper.AddConfigPath("./config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't load config file")
	}
}

func setupInflux() {
	// config vals
	url := viper.GetString("influx.url")
	token := viper.GetString("influx.token")
	org := viper.GetString("influx.org")
	bucket := viper.GetString("influx.bucket")
	// setup influx
	influxClient = influxdb2.NewClient(url, token)
	influxWriter = influxClient.WriteAPI(org, bucket)
}

func newEvent(w http.ResponseWriter, r *http.Request) {
	return
}

func main() {
	// configure logger and get the config
	configLogger()
	getConfig()
	// setup influx using the config values
	setupInflux()
	// setup and start gorilla
	r := mux.NewRouter()
	r.HandleFunc("/event", newEvent).Methods("POST")
	serverAddr := viper.GetString("server.address")
	if viper.GetBool("server.useSSL") {
		certPath := viper.GetString("server.certPath")
		keyPath := viper.GetString("server.keyPath")
		log.Info().Str("keypath", keyPath).Str("certPath", certPath).Msg("starting API server using SSL")
		err := http.ListenAndServeTLS(serverAddr, certPath, keyPath, r)
		log.Fatal().Err(err).Msg("API server stopped")
	} else {
		log.Warn().Msg("starting API server without SSL")
		err := http.ListenAndServe(serverAddr, r)
		log.Fatal().Err(err).Msg("API server stopped")
	}
}
