package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	g "github.com/soniah/gosnmp"
)

var drops, sent int64
var total, cores int
var constant bool
var host string

func init() {
	flag.IntVar(&total, "total", 50000, "total number of traps to send")
	flag.IntVar(&cores, "workers", runtime.NumCPU(), "number of worker routines")
	flag.BoolVar(&constant, "constant", false, "don't stop sending")
	flag.StringVar(&host, "host", "127.0.0.1", "ip/hostname of the trap receiver")
}

func main() {
	flag.Parse()
	st := time.Now()

	// Default is a pointer to a GoSNMP struct that contains sensible defaults

	count := total / cores
	var wg sync.WaitGroup

	fmt.Printf("constant: %v, workers %v, count %v, total %v\n", constant, cores, count, total)
	for i := 0; i < cores; i++ {
		wg.Add(1)
		work(count, &wg)
	}
	wg.Wait()
	dur := time.Since(st)
	tps := float64(total) / dur.Seconds()
	fmt.Printf("sent %v traps in %v %.2f tps, drops: %v\n", sent, dur, tps, drops)

}
func work(count int, wg *sync.WaitGroup) {
	defer wg.Done()
	snmp := &g.GoSNMP{
		Target:  host,
		Port:    162,
		Version: g.Version1,
		Timeout: time.Second * 5,
		Retries: 1,
	}
	err := snmp.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer snmp.Conn.Close()

	if !constant {
		for i := 0; i < count; i++ {
			sendTrap(snmp)
		}
	} else {
		for {
			sendTrap(snmp)
		}
	}

}

func sendTrap(snmp *g.GoSNMP) error {

	r := 1
	//add some randomnesssssss
	if n := rand.Intn(10); n > 1 {
		r = n
	}
	pdus := make([]g.SnmpPDU, r)

	for i := 0; i < r; i++ {
		pdu := g.SnmpPDU{
			Name:  ".1.3.6.1.2.1.1.6." + strconv.Itoa(i),
			Type:  g.OctetString,
			Value: "Your mom",
		}
		pdus[i] = pdu
	}

	trap := g.SnmpTrap{
		Variables:    pdus,
		Enterprise:   ".1.3.6.1.6.3.1.1.5.1",
		AgentAddress: "127.0.0.1",
		GenericTrap:  0,
		SpecificTrap: 0,
		Timestamp:    300,
	}

	_, err := snmp.SendTrap(trap)

	if err != nil {
		atomic.AddInt64(&drops, 1)
		fmt.Println(err)
		return err
	}
	atomic.AddInt64(&sent, 1)
	return nil
}
