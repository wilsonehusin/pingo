package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wilsonehusin/pingo"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})

	log.SetOutput(os.Stderr)

	log.SetLevel(log.InfoLevel)
}

func main() {
	ipAddr := flag.String("ipaddr", "1.1.1.1", "IP address to target")
	strData := flag.String("strdata", "pingo!", "Data to be sent and expect back")
	interval := flag.Duration("interval", time.Duration(5)*time.Second, "Time (duration) between pings")
	timeout := flag.Duration("timeout", time.Duration(0), "Length of timeout between send and receive (0 = no timeout)")

	flag.Parse()

	targetIP := net.UDPAddr{IP: net.ParseIP(*ipAddr)}
	dataToSend := []byte(*strData)

	userStop := make(chan os.Signal)
	loopStop := make(chan bool)

	signal.Notify(userStop, os.Interrupt)

	go func() {
		<-userStop
		loopStop <- true
	}()

	pingo.SendIndefinitely(targetIP, dataToSend, *interval, *timeout, loopStop)
}
