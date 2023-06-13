package pbft

import (
	"fmt"
	"sync"
)

func genPBFTSynchronize(numNodes int, data string, clientAddr string) float64 {

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
		p := NewPBFT(nodeID, nodeTable[nodeID], nodeTable, numNodes)
		go p.tcpListen(ready) // Pass the 'ready' channel to tcpListen
	}

	for i := 0; i < numNodes; i++ {
		<-ready // Wait for all nodes to signal readiness
	}

	// Now all nodes are ready, initiate the client node
	println("initiating client...")
	wg.Add(1) // We are adding 1 goroutine we want to wait for
	go func() {
		elapsedTime = clientSendMessageAndListen(clientAddr, nodeTable, data, numNodes)
		wg.Done() // Signal that the goroutine is finished
	}()
	wg.Wait() // Wait until all goroutines have finished
	return elapsedTime
}
