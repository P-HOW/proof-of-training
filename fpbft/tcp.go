package pbft

import (
	"log"
	"net"
)

// TCP send messages
func tcpDial(context []byte, addr string, bandwidthLimit int, latency float64) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("connect error", err)
		return
	}
	defer conn.Close()

	tw := throttledWriter{w: conn, bandwidthLimit: bandwidthLimit}

	// Apply latency before writing
	applyLatency(latency)

	if _, err := tw.Write(context); err != nil {
		log.Println(err)
		return
	}

}
