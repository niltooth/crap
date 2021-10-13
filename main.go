package main

import (
	"flag"
	"os"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var configPath, logLevel string
var natsErrs, received, drops int64

func init() {
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
}

func main() {
	flag.Parse()

	//logger, _ := zap.NewProduction()
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg, err := getConfig(configPath)
	if err != nil {
		logger.Fatal(err.Error())
	}

	nargs := []nats.Option{}
	if cfg.Nats.CredsFile != "" {
		nargs = append(nargs, nats.UserCredentials(cfg.Nats.CredsFile))
	}
	if cfg.Nats.RootCa != "" {
		nargs = append(nargs, nats.RootCAs(cfg.Nats.RootCa))
	}
	nargs = append(nargs, nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
		atomic.AddInt64(&natsErrs, 1)
	}))
	nargs = append(nargs, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		logger.Info("nats disconnect",
			zap.String("error", err.Error()),
		)
		atomic.AddInt64(&natsErrs, 1)
	}))
	nargs = append(nargs, nats.ReconnectHandler(func(conn *nats.Conn) {
		logger.Info("nats reconnect")
	}))
	nc, err := nats.Connect(cfg.Nats.Url, nargs...)
	if err != nil {
		logger.Fatal(err.Error())
	}

	tp, th, err := NewTrapServer(cfg, nc, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if cfg.Stats.Subject != "" {
		go th.StartStats(cfg)
	}
	logger.Info("Listening for traps", zap.Int("port", cfg.Port))
	tp.ListenAndServe(th)

}

func getConfig(configPath string) (*Config, error) {
	config := &Config{}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, config); err != nil {
		return nil, err
	}
	return config, nil
}
