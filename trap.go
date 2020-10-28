package main

import (
	"github.com/dev-mull/pgbuffer"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	//g "github.com/soniah/gosnmp"
	snmplib "github.com/deejross/go-snmplib"
	//log "github.com/sirupsen/logrus"
)

type snmpHandler struct {
	buffer      *pgbuffer.Buffer
	destination chan *Trap
}

func (h *snmpHandler) OnError(addr net.Addr, err error) {

	atomic.AddInt64(&drops, 1)
	log.Println(addr.String(), err)

}

func (h *snmpHandler) OnTrap(addr net.Addr, trap snmplib.Trap) {
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
	if trapoid, ok := t.VarBinds[".1.3.6.1.6.3.1.1.4.1.0"].(snmplib.Oid); ok {
		var s []string
		for _, i := range trapoid {
			text := strconv.Itoa(i)
			s = append(s, text)
		}
		t.TrapOid = "." + strings.Join(s, ".")
		t.VarBinds[".1.3.6.1.6.3.1.1.4.1.0"] = t.TrapOid

	} else if trap.OID.String() != "" && trap.OID.String() != "." {
		t.TrapOid = trap.OID.String()
		t.VarBinds[".1.3.6.1.6.3.1.1.4.1.0"] = t.TrapOid

	}
	h.destination <- &t
}
