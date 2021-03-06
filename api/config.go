package main

type Config struct {
	Server struct {
		Address  string `yaml:"address"`
		UseSSL   bool   `yaml:"useSSL"`
		CertPath string `yaml:"certPath"`
		KeyPath  string `yaml:"keyPath"`
	} `yaml:"server"`
	Timescale struct {
		Url      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"timescale"`
}
