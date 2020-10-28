package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"os"
	"sync/atomic"
	"time"

	snmplib "github.com/deejross/go-snmplib"
	"github.com/dev-mull/pgbuffer"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
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
var conn *nats.Conn

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
	dest := make(chan *Trap, 10000)
	stop := make(chan bool, 2)
	tp, th, err := NewTrapServer(cfg, dest)
	if err != nil {
		log.Fatal(err)
	}
	go RunServer(cfg, dest, stop)
	defer func() { stop <- true }()
	log.Infof("Listening for traps on port %v\n", cfg.Port)
	tp.ListenAndServe(th)

}
func NewTrapServer(cfg *Config, dst chan *Trap) (*snmplib.TrapServer, *snmpHandler, error) {
	server, err := snmplib.NewTrapServer("0.0.0.0", cfg.Port)
	if err != nil {
		return nil, nil, err
	}
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

	return &server, &snmpHandler{destination: dst}, nil
}

func RunServer(cfg *Config, ch chan *Trap, stop chan bool) error {
	var err error
	if cfg.Mode == "hybrid" || cfg.Mode == "db-only" {
		db, err = sql.Open("postgres", cfg.DB)
		if err != nil {
			return err
		}
		if err := db.Ping(); err != nil {
			return err
		}
		buff = pgbuffer.NewBuffer(db, cfg.Buffer, log)
		defer db.Close()
		go buff.Run()
	}
	if cfg.Mode == "hybrid" || cfg.Mode == "nats-only" {
		conn, err = nats.Connect(cfg.Nats)
		if err != nil {
			return err
		}
		defer conn.Close()
	}
	//optimized loops so we are not checking the same thing over and over at runtime
	go StartStats()
	switch cfg.Mode {
	case "hybrid":
		for {
			select {
			case <-stop:
				break
			case t := <-ch:
				b, _ := json.Marshal(t)
				conn.Publish(cfg.Subject, b)
				buff.Write("traps", t.Time.Format(time.RFC3339), string(b))
			}
		}
	case "db-only":
		for {
			select {
			case <-stop:
				break
			case t := <-ch:
				b, _ := json.Marshal(t)
				buff.Write("traps", t.Time.Format(time.RFC3339), string(b))
			}
		}
	case "nats-only":
		for {
			select {
			case <-stop:
				break
			case t := <-ch:
				b, _ := json.Marshal(t)
				conn.Publish(cfg.Subject, b)
			}
		}
	}

	return nil
}

func StartStats() {
	if cfg.StatsInterval != 0 {
		t := time.NewTicker(cfg.StatsInterval)
		for {
			select {
			case <-t.C:
				count := atomic.LoadInt64(&received)
				dropCount := atomic.LoadInt64(&drops)
				if cfg.Mode == "hybrid" || cfg.Mode == "db-only" {
					db.Exec(`insert into trap_stats (time,id,received,drops) values ($1,$2,$3,$4)`, time.Now(), hostname, count, dropCount)
				} else {
					s := Stat{
						Time:     time.Now(),
						Hostname: hostname,
						Drops:    dropCount,
						Count:    count,
					}
					b, _ := json.Marshal(s)
					conn.Publish(cfg.StatsSubject, b)
				}

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
