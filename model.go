package main

import (
	"time"

	"github.com/dev-mull/pgbuffer"
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

type Config struct {
	Port          int              `yaml:"port"`
	Users         []User           `yaml:"users"`
	DB            string           `yaml:"db"`
	Buffer        *pgbuffer.Config `yaml:"buffer"`
	StatsInterval time.Duration    `yaml:"stats-interval"`
}
type User struct {
	User       string `yaml:"user"`
	AuthAlg    string `yaml:"auth-alg"`
	AuthPasswd string `yaml:"auth-passwd"`
	PrivAlg    string `yaml:"priv-alg"`
	PrivPasswd string `yaml:"priv-passwd"`
}
