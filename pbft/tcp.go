package pbft

import (
	"io/ioutil"
	"log"
	"net"
)

// TCP listening from clinet side
func clientTcpListen(clientAddr string, numNodes int) {
	listen, err := net.Listen("tcp", clientAddr)
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
		if count == numNodes {
			break
		}
	}

}

// TCP listening from node side
func (p *pbft) tcpListen(ready chan<- bool) {
	//println("starting")
	listen, err := net.Listen("tcp", p.node.addr)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("Node listening starts, address：%s\n", p.node.addr)
	defer listen.Close()
	ready <- true // Signal that the server is ready
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Panic(err)
		}
		b, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Panic(err)
		}
		p.handleRequest(b)
	}

}

// TCP send messages
func tcpDial(context []byte, addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("connect error", err)
		return
	}

	_, err = conn.Write(context)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
}
