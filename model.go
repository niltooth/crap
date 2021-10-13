package main

import (
	"time"
)

type Config struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Users    []User `yaml:"users"`
	Encoding string `yaml:"encoding"`
	Nats     struct {
		Url       string `yaml:"url"`
		CredsFile string `yaml:"creds-file"`
		Subject   string `yaml:"subject"`
		RootCa    string `yaml:"root-ca"`
		Jetstream bool   `yaml:"jetstream"`
	} `yaml:"nats"`
	Stats struct {
		Subject  string        `yaml:"subject"`
		Interval time.Duration `yaml:"interval"`
	} `yaml:"stats"`
}

type User struct {
	User       string `yaml:"user"`
	AuthAlg    string `yaml:"auth-alg"`
	AuthPasswd string `yaml:"auth-passwd"`
	PrivAlg    string `yaml:"priv-alg"`
	PrivPasswd string `yaml:"priv-passwd"`
}
