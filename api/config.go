package main

type Config struct {
	Server struct {
		Address  string `yaml:"address"`
		UseSSL   bool   `yaml:"useSSL"`
		CertPath string `yaml:"certPath"`
		KeyPath  string `yaml:"keyPath"`
	} `yaml:"server"`
	Influx struct {
		Url    string `yaml:"url"`
		Token  string `yaml:"token"`
		Org    string `yaml:"org"`
		Bucket string `yaml:"bucket"`
	} `yaml:"influx"`
}
