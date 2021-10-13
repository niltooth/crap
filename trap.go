package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/deejross/go-snmplib"
	"github.com/dev-mull/crap/pb"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewTrapServer(cfg *Config, nc *nats.Conn, logger *zap.Logger) (*snmplib.TrapServer,
	*Handler, error) {
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

	h, err := NewHandler(nc, cfg, logger)
	return &server, h, err
}

type Handler struct {
	c      *nats.Conn
	js     nats.JetStreamContext
	cfg    *Config
	logger *zap.Logger
}

func NewHandler(nc *nats.Conn, cfg *Config, logger *zap.Logger) (*Handler, error) {
	h := &Handler{
		c:      nc,
		cfg:    cfg,
		logger: logger,
	}
	if cfg.Nats.Jetstream {
		var err error
		if h.js, err = nc.JetStream(); err != nil {
			return nil, err
		}
	}
	return h, nil
}

//I'm not sure on the use of this oid to represent the trap oid...
const trapOid = ".1.3.6.1.6.3.1.1.4.1.0"

//TODO: add optional jetstream context..
//IDEA: optional subject based on trapoid
func (h *Handler) OnTrap(addr net.Addr, trap snmplib.Trap) {
	go func(addr net.Addr, trap snmplib.Trap) {
		atomic.AddInt64(&received, 1)

		//gotta do a little funny stuff here.
		//probably should write my own protobuf capable snmp lib
		vb := make(map[string]interface{})
		for k, v := range trap.VarBinds {
			var nv interface{}
			switch v.(type) {
			case snmplib.Oid:
				nv = v.(snmplib.Oid).String()
			case time.Duration:
				nv = v.(time.Duration).String()
			default:
				nv = v
			}
			vb[k] = nv
		}
		nvb, err := structpb.NewStruct(vb)
		if err != nil {
			h.logger.Error("failed to map varbinds", zap.String("error", err.Error()))
			atomic.AddInt64(&drops, 1)
			return
		}
		tt := &pb.Trap{
			Time:      timestamppb.Now(),
			Address:   trap.Address,
			Version:   int32(trap.Version),
			TrapType:  int32(trap.TrapType),
			Community: trap.Community,
			User:      trap.Username,
			VarBinds:  nvb,
		}

		if o, ok := trap.VarBinds[trapOid].(snmplib.Oid); ok {
			s := make([]string, len(o))
			for i, v := range o {
				s[i] = strconv.Itoa(v)
			}
			tt.TrapOid = "." + strings.Join(s, ".")

		} else if s := trap.OID.String(); s != "" && s != "." {
			tt.TrapOid = s
		}

		sub := h.cfg.Nats.Subject + "." +
			strings.ReplaceAll(trap.Community, ".", "_") + "." +
			strings.ReplaceAll(tt.TrapOid, ".", "-")

		if err := h.Publish(sub, tt); err != nil {
			h.logger.Error("failed publish",
				zap.String("error", err.Error()),
				zap.String("sub", sub),
			)
			atomic.AddInt64(&drops, 1)
		}
	}(addr, trap)
}

func (h *Handler) Publish(subject string, m proto.Message) error {
	var b []byte
	var err error
	switch h.cfg.Encoding {
	case "json":
		b, err = protojson.MarshalOptions{UseProtoNames: true}.Marshal(m)
	default:
		b, err = proto.Marshal(m)
	}
	if err != nil {
		return err
	}
	return h.c.Publish(subject, b)
}

func (h *Handler) OnError(addr net.Addr, err error) {
	h.logger.Error("failed to handle trap",
		zap.String("error", err.Error()),
		zap.String("addr", addr.String()),
	)
	atomic.AddInt64(&drops, 1)
}

func (h *Handler) StartStats(cfg *Config) {
	if cfg.Stats.Interval != 0 {
		hostname, _ := os.Hostname() // good enough
		t := time.NewTicker(cfg.Stats.Interval)
		for {
			select {
			case <-t.C:
				s := &pb.Stat{
					Time:       timestamppb.Now(),
					Hostname:   hostname,
					Drops:      atomic.LoadInt64(&drops),
					Received:   atomic.LoadInt64(&received),
					NatsErrors: atomic.LoadInt64(&natsErrs),
				}
				h.logger.Info("stats",
					zap.Int64("drops", s.Drops),
					zap.Int64("received", s.Received),
					zap.Int64("nats-errors", s.NatsErrors),
				)
				h.Publish(cfg.Stats.Subject, s)
			}

		}
	}
}
