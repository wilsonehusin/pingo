package pingo

import (
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func sendThroughConnection(connection *icmp.PacketConn, target net.UDPAddr, data []byte) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: data,
		},
	}

	encodedMessage, err := message.Marshal(nil)
	if err != nil {
		log.Panic(err)
	}
	sendTime := time.Now()
	if _, err := connection.WriteTo(encodedMessage, &target); err != nil {
		log.Error(err)
		return
	}

	encodedResponse := make([]byte, 1500)
	n, peer, err := connection.ReadFrom(encodedResponse)
	if err != nil {
		log.Error(err)
		return
	}
	recvDur := time.Since(sendTime)

	response, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), encodedResponse[:n])
	if err != nil {
		log.Error(err)
		return
	}

	switch response.Type {
	case ipv4.ICMPTypeEchoReply:
		responseData := response.Body.(*icmp.Echo).Data
		log.Infof("received ICMP echo reply from %v: %v (took %v)", peer, string(responseData), recvDur)
	default:
		log.Warn("received %+v; expected echo response", response)
	}
}

func Send(target net.UDPAddr, data []byte, timeout time.Duration) {
	connection, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Panic(err)
	}
	defer connection.Close()

	if timeout != time.Duration(0) {
		log.Trace("timeout is set to ", timeout)
		connection.SetDeadline(time.Now().Add(timeout))
	}
	sendThroughConnection(connection, target, data)
}

func SendIndefinitely(target net.UDPAddr, data []byte, interval time.Duration, timeout time.Duration, stop chan bool) {
	connection, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Panic(err)
	}
	defer connection.Close()

	if timeout != time.Duration(0) {
		log.Trace("timeout is set to ", timeout)
		connection.SetDeadline(time.Now().Add(timeout))
	}

	ticker := time.NewTicker(interval)
	log.Info("sending first ping in ", interval)
	for {
		select {
		case <-stop:
			log.Trace("done sending pings")
			return
		case c := <-ticker.C:
			log.Trace("sending at ", c)
			go sendThroughConnection(connection, target, data)
		}
	}
}
