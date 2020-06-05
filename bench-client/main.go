package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	g "github.com/soniah/gosnmp"
)

var drops, sent int64
var total, cores, count int
var constant bool

func init() {
	flag.IntVar(&total, "total", 50000, "total number of traps to send")
	flag.IntVar(&cores, "workers", runtime.NumCPU(), "number of worker routines")
	flag.BoolVar(&constant, "constant", false, "don't stop sending")
}

func main() {
	flag.Parse()
	st := time.Now()

	// Default is a pointer to a GoSNMP struct that contains sensible defaults

	// eg port 161, community public, etc
	var wg sync.WaitGroup
	count = total / cores
	fmt.Printf("constant: %v, workers %v, count %v, total %v\n", constant, cores, count, total)
	for i := 0; i < cores; i++ {
		wg.Add(1)
		work(&wg)
	}
	wg.Wait()
	dur := time.Since(st)
	tps := float64(total) / dur.Seconds()
	fmt.Printf("sent %v traps in %v %v tps drops: %v\n", sent, dur, tps, drops)

}
func work(wg *sync.WaitGroup) {
	defer wg.Done()
	if !constant {

		for i := 0; i < count; i++ {
			sendTrap()
		}
	} else {
		for {
			sendTrap()
		}
	}

}

func sendTrap() error {
	snmp := &g.GoSNMP{
		Target:  "127.0.0.1",
		Port:    162,
		Version: g.Version1,
		Timeout: time.Second * 5,
		Retries: 1,
	}
	err := snmp.Connect()
	if err != nil {
		atomic.AddInt64(&drops, 1)
		return err
	}
	defer snmp.Conn.Close()

	pdu := g.SnmpPDU{
		Name:  "1.3.6.1.2.1.1.6",
		Type:  g.OctetString,
		Value: "Your mom",
	}

	trap := g.SnmpTrap{
		Variables:    []g.SnmpPDU{pdu},
		Enterprise:   ".1.3.6.1.6.3.1.1.5.1",
		AgentAddress: "127.0.0.1",
		GenericTrap:  0,
		SpecificTrap: 0,
		Timestamp:    300,
	}

	_, err = snmp.SendTrap(trap)

	if err != nil {
		atomic.AddInt64(&drops, 1)
		return err
	}
	atomic.AddInt64(&sent, 1)
	return nil
}
