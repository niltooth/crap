package main

import (
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/deejross/go-snmplib"
	"github.com/nats-io/nats.go"
)

func NewTrapServer(cfg *Config, nc *nats.Conn, subject string) (*snmplib.TrapServer, *Handler, error) {
	server, err := snmplib.NewTrapServer(cfg.Address, cfg.Port)
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

	return &server, NewHandler(nc, subject), nil
}

type Handler struct {
	c       *nats.EncodedConn
	subject string
}

func NewHandler(nc *nats.Conn, subject string) *Handler {
	enc, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	return &Handler{
		enc, subject,
	}
}

//I'm not sure on the use of this oid to represent the trap oid...
const trapOid = ".1.3.6.1.6.3.1.1.4.1.0"

//TODO: Convert to use optional protobuf... probably need to bring in parts of the snmp lib..
//TODO: add optional jetstream context..
//IDEA: optional subject based on trapoid
func (h *Handler) OnTrap(addr net.Addr, trap snmplib.Trap) {
	go func(addr net.Addr, trap snmplib.Trap) {
		atomic.AddInt64(&received, 1)

		t := Trap{
			Time:      time.Now(),
			VarBinds:  trap.VarBinds,
			Address:   trap.Address,
			Version:   trap.Version,
			TrapType:  trap.TrapType,
			Community: trap.Community,
			Username:  trap.Username,
		}

		if o, ok := t.VarBinds[trapOid].(snmplib.Oid); ok {
			s := make([]string, len(o))
			for i, v := range o {
				s[i] = strconv.Itoa(v)
			}
			t.TrapOid = "." + strings.Join(s, ".")
			t.VarBinds[trapOid] = t.TrapOid

		} else if s := trap.OID.String(); s != "" && s != "." {
			t.TrapOid = s
			t.VarBinds[trapOid] = t.TrapOid
		}

		h.c.Publish(h.subject, t)
	}(addr, trap)
}

func (h *Handler) OnError(addr net.Addr, err error) {
	atomic.AddInt64(&drops, 1)
}

func (h *Handler) StartStats(cfg *Config) {
	if cfg.Stats.Interval != 0 {
		t := time.NewTicker(cfg.Stats.Interval)
		for {
			select {
			case <-t.C:
				count := atomic.LoadInt64(&received)
				dropCount := atomic.LoadInt64(&drops)
				s := Stat{
					Time:     time.Now(),
					Hostname: hostname,
					Drops:    dropCount,
					Count:    count,
				}
				h.c.Publish(cfg.Stats.Subject, s)
			}

		}
	}
}
