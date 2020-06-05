package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	snmplib "github.com/deejross/go-snmplib"
	"github.com/dev-mull/pgbuffer"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var configPath, logLevel string
var buff *pgbuffer.Buffer
var log = logrus.New()
var cfg *Config
var received, drops int64
var hostname string
var err error
var db *sql.DB

func init() {
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.StringVar(&logLevel, "log-level", "info", "log level")
}

func main() {
	flag.Parse()

	log.Out = os.Stdout
	switch logLevel {
	case "trace":
		log.Level = logrus.TraceLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "info":
		log.Level = logrus.InfoLevel
	case "debug":
		log.Level = logrus.DebugLevel
	default: //just in case someone puts in something invalid
		log.Level = logrus.InfoLevel
	}
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err = NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("postgres", cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Debug("connected to db")

	buff = pgbuffer.NewBuffer(db, cfg.Buffer, log)
	fmt.Printf("%+v\n", cfg.Buffer)
	go buff.Run()

	server, err := snmplib.NewTrapServer("0.0.0.0", cfg.Port)
	for _, u := range cfg.Users {
		user := snmplib.V3user{
			User:    u.User,
			AuthAlg: u.AuthAlg,
			AuthPwd: u.AuthPasswd,
			PrivAlg: u.PrivAlg,
			PrivPwd: u.PrivPasswd,
		}
		server.Users = append(server.Users, user)
	}

	if err != nil {
		log.Fatalln(err)
	}

	log.Infof("Listening for traps on port %v\n", cfg.Port)
	go StartStats()

	server.ListenAndServe(snmpHandler{})
}

func StartStats() {
	if cfg.StatsInterval != 0 {
		t := time.NewTicker(cfg.StatsInterval)
		for {
			select {
			case <-t.C:
				count := atomic.LoadInt64(&received)
				dropCount := atomic.LoadInt64(&drops)
				db.Exec(`insert into trap_stats (time,id,received,drops) values ($1,$2,$3,$4)`, time.Now(), hostname, count, dropCount)
			}

		}
	}
}

func NewConfig(configPath string) (*Config, error) {
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
