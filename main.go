package main

import (
	"flag"
	"log"
	"os"

	snmplib "github.com/deejross/go-snmplib"
	"gopkg.in/yaml.v2"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
}

func main() {
	flag.Parse()
	cfg, err := NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
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

	log.Printf("Listening for traps on port %v\n", cfg.Port)

	server.ListenAndServe(snmpHandler{})
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
