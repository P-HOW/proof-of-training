package pbft

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

func clientSendMessageAndListen(clientAddr string, nodeTable nodeTable, data string, numNodes int) float64 {
	var wg sync.WaitGroup

	//Start local monitoring of the client (mainly used to receive reply information from nodes).
	go func() {
		defer wg.Done()
		clientTcpListen(clientAddr, numNodes)
	}()

	wg.Add(1) // Expect numNodes number of replies

	r := new(Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = clientAddr
	r.Message.ID = getRandom()
	//The message content is the user's input
	r.Message.Content = strings.TrimSpace(data)
	br, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(br))
	content := jointMessage(cRequest, br)
	currentTime := time.Now()
	//N0 is the primary node, and the request information is sent directly to N0 by default
	tcpDial(content, nodeTable["N0"])

	wg.Wait() // Wait for all the replies before proceeding

	return time.Since(currentTime).Seconds()

}

// 返回一个十位数的随机数，作为msgid
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}
