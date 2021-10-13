package main

import (
	"flag"
	"os"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var configPath, logLevel string
var received, drops int64

func init() {
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
}

func main() {
	flag.Parse()

	cfg, err := getConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	nargs := []nats.Option{}
	if cfg.Nats.CredsFile != "" {
		nargs = append(nargs, nats.UserCredentials(cfg.Nats.CredsFile))
	}
	if cfg.Nats.RootCa != "" {
		nargs = append(nargs, nats.RootCAs(cfg.Nats.RootCa))
	}
	nc, err := nats.Connect(cfg.Nats.Url, nargs...)
	if err != nil {
		log.Fatal(err)
	}

	tp, th, err := NewTrapServer(cfg, nc, cfg.Nats.Subject)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Stats.Subject != "" {
		go th.StartStats(cfg)
	}
	log.Infof("Listening for traps on port %v\n", cfg.Port)
	tp.ListenAndServe(th)

}

func getConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
