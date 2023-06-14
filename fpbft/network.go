package fpbft

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

func genPBFTSynchronize(numNodes int, data string, clientAddr string, bandwidth float64, latency float64) float64 {

	var wg sync.WaitGroup
	var elapsedTime float64

	genRsaKeys(numNodes)

	nodeTable := make(map[string]string) // Initialize the map
	for i := 0; i < numNodes; i++ {
		nodeID := fmt.Sprintf("N%d", i)
		nodeTable[nodeID] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
	}

	ready := make(chan bool, numNodes) // Create a buffered channel
	for i := 0; i < numNodes; i++ {
		nodeID := fmt.Sprintf("N%d", i)
		p := NewPBFT(nodeID, nodeTable[nodeID], nodeTable, numNodes, bandwidth, latency)
		go p.tcpListen(ready) // Pass the 'ready' channel to tcpListen
	}

	for i := 0; i < numNodes; i++ {
		<-ready // Wait for all nodes to signal readiness
	}

	// Now all nodes are ready, initiate the client node
	println("initiating client...")
	myClient := client{
		clientAddr: clientAddr,
		index:      1,
		bandwidth:  bandwidth,
		latency:    latency,
	}
	wg.Add(1) // We are adding 1 goroutine we want to wait for
	go func() {
		elapsedTime = myClient.ClientSendMessageAndListen(nodeTable, data, numNodes)
		wg.Done() // Signal that the goroutine is finished
	}()
	wg.Wait() // Wait until all goroutines have finished
	return elapsedTime
}

func applyLatency(t float64) {
	r := rand.Float64()            // generates a random float between 0.0 and 1.0
	latency := 0.1*t + r*(t-0.1*t) // calculate latency in range of 0.1t to t
	time.Sleep(time.Duration(latency) * time.Millisecond)
}

type throttledWriter struct {
	w              io.Writer
	bandwidthLimit int
}

func (tw *throttledWriter) Write(p []byte) (n int, err error) {
	chunkSize := tw.bandwidthLimit / 10 // Adjust this as per your requirements
	kk := 0
	for len(p) > 0 {
		//println("writing the " + strconv.Itoa(kk) + "th chunk....")
		kk++
		time.Sleep(time.Second / 10) // Simulate bandwidth limit
		chunk := p
		//println(len(chunk))
		if len(chunk) > chunkSize {
			chunk = chunk[:chunkSize]
		}
		n, err = tw.w.Write(chunk)

		// Check if an error occurred
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				// If it's a temporary error, we just continue to the next iteration
				continue
			} else {
				// For other types of errors, we return from the function
				return n, err
			}
		}
		// Move to the next chunk
		p = p[n:]
	}
	return
}
