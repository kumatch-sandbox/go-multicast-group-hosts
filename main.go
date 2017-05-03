package main

import (
	"bytes"
	"log"
	"net"
	"time"
)

const (
	multicastAddress  = "224.0.0.151:9999"
	maxSize           = 8192
	heartBeatProtocol = "kumatch-sandbox/go-multicast-1"
)

func main() {
	hosts := newHosts()

	go heartBeat(multicastAddress, heartBeatProtocol)
	go check(hosts)
	serve(hosts, multicastAddress, heartBeatProtocol)
}

func heartBeat(address, protocol string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn.Write([]byte(protocol))
		}
	}
}

func check(hosts *hosts) {
	ticker := time.NewTicker((1000 / 60) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-hosts.updated:
			hosts.Display()
		case <-ticker.C:
			hosts.Check()
		}
	}
}

func serve(hosts *hosts, address, protocol string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, udpAddr)
	conn.SetReadBuffer(maxSize)
	for {
		b := make([]byte, maxSize)
		n, srcAddr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		if n > 0 && protocol == decode(b) {
			hosts.Add(srcAddr.IP.String())
		}
	}
}

func decode(b []byte) string {
	n := bytes.IndexByte(b, 0)
	return string(b[:n])
}
