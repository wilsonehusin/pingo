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

func summarize(nums []float64) (int, float64) {
	count := len(nums)
	sum := 0.0
	for _, num := range nums {
		sum += num
	}
	return count, (sum / float64(count))
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

	results := pingo.SendIndefinitely(targetIP, dataToSend, *interval, *timeout, loopStop)
	var durations []float64

	for _, result := range results {
		durations = append(durations, float64(result/time.Microsecond)/1000)
	}
	iteration, average := summarize(durations)
	log.Info("average: ", average, " milliseconds, over ", iteration, " iterations")
}
