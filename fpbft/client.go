package fpbft

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"
)

type client struct {
	clientAddr string
	index      int //client ID for convenience purposes
	bandwidth  float64
	latency    float64
}

func (c *client) ClientSendMessageAndListen(nodeTable nodeTable, data string, numNodes int) float64 {
	var wg sync.WaitGroup

	//Start local monitoring of the client (mainly used to receive reply information from nodes).
	go func() {
		defer wg.Done()
		c.clientTcpListen(numNodes)
	}()

	wg.Add(1) // Expect numNodes number of replies

	r := new(Request)
	r.Timestamp = time.Now().UnixNano()
	r.ClientAddr = c.clientAddr
	r.Message.ID = getRandom()
	//The message content is the user's input
	r.Message.Content = strings.TrimSpace(data)
	br, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Println(string(br))
	content := jointMessage(cRequest, br)
	currentTime := time.Now()
	//N0 is the primary node, and the request information is sent directly to N0 by default
	tcpDial(content, nodeTable["N0"], int(c.bandwidth*1024*1024/8), c.latency)

	wg.Wait() // Wait for all the replies before proceeding

	return time.Since(currentTime).Seconds()

}

// TCP listening from clinet side
func (c *client) clientTcpListen(numNodes int) {
	listen, err := net.Listen("tcp", c.clientAddr)
	if err != nil {
		log.Panic(err)
	}
	defer listen.Close()
	count := 0
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Panic(err)
		}
		b, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Panic(err)
		}
		//fmt.Println("client received" + string(b))
		_ = b
		count++
		if count > numNodes/3*2 {
			break
		}
	}
}

// Returns a ten-digit random number as msgid
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
