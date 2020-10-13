package pingo

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ResponseError struct {
	Response icmp.Message
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("received unexpected response type: %+v", e.Response)
}

func sendThroughConnection(connection *icmp.PacketConn, target net.UDPAddr, data []byte) (time.Duration, error) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: data,
		},
	}

	encodedMessage, err := message.Marshal(nil)
	if err != nil {
		return time.Duration(0), err
	}
	sendTime := time.Now()
	if _, err := connection.WriteTo(encodedMessage, &target); err != nil {
		return time.Duration(0), err
	}

	encodedResponse := make([]byte, 1500)
	n, _, err := connection.ReadFrom(encodedResponse)
	if err != nil {
		return time.Duration(0), err
	}
	recvDur := time.Since(sendTime)

	response, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), encodedResponse[:n])
	if err != nil {
		return time.Duration(0), err
	}

	switch response.Type {
	case ipv4.ICMPTypeEchoReply:
		return recvDur, nil
	default:
		return recvDur, &ResponseError{*response}
	}
}

type PackageSlip struct {
	Target   net.UDPAddr
	Data     []byte
	Timeout  time.Duration
	Interval time.Duration
}

func Send(slip PackageSlip) (time.Duration, error) {
	connection, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Panic(err)
	}
	defer connection.Close()

	if slip.Timeout != time.Duration(0) {
		log.Trace("timeout is set to ", slip.Timeout)
		connection.SetDeadline(time.Now().Add(slip.Timeout))
	}
	return sendThroughConnection(connection, slip.Target, slip.Data)
}

func SendIndefinitely(ctx context.Context, slip PackageSlip) []time.Duration {
	connection, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Panic(err)
	}
	defer connection.Close()

	if slip.Timeout != time.Duration(0) {
		log.Trace("timeout is set to ", slip.Timeout)
	}

	ticker := time.NewTicker(slip.Interval)
	log.Info("sending first ping in ", slip.Interval)

	var responseDurations []time.Duration
	for {
		select {
		case <-ctx.Done():
			log.Info("done sending pings")
			return responseDurations
		case c := <-ticker.C:
			if slip.Timeout != time.Duration(0) {
				connection.SetDeadline(time.Now().Add(slip.Timeout))
			}
			log.Trace("sending at ", c)
			duration, err := sendThroughConnection(connection, slip.Target, slip.Data)
			if err != nil {
				log.Error(err)
			}
			log.Info("received response within ", duration)
			responseDurations = append(responseDurations, duration)
		}
	}
}
