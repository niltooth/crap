package main

import (
	"time"
)

type Trap struct {
	Time      time.Time              `json:"time"`
	Version   int                    `json:"version"`
	TrapType  int                    `json:"trap_type,omitempty"`
	Other     interface{}            `json:"other,omitempty"`
	Community string                 `json:"community,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Address   string                 `json:"address"`
	TrapOid   string                 `json:"trap_oid"`
	VarBinds  map[string]interface{} `json:"var_binds"`
}
type Stat struct {
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Count    int64     `json:"count"`
	Drops    int64     `json:"drops"`
}

type Config struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	Users   []User `yaml:"users"`
	Nats    struct {
		Url       string `yaml:"url"`
		CredsFile string `yaml:"creds-file"`
		Subject   string `yaml:"subject"`
		RootCa    string `yaml:"root-ca"`
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
