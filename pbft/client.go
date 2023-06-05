package pbft

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

func clientSendMessageAndListen(clientAddr string, nodeTable nodeTable) {
	//Start local monitoring of the client (mainly used to receive reply information from nodes).
	go clientTcpListen(clientAddr)
	fmt.Printf("Client starts listening, address: %s\n", clientAddr)

	fmt.Println(" ---------------------------------------------------------------------------------")
	fmt.Println("|  Entered PBFT test Demo client, please start all nodes before sending messages! :)  |")
	fmt.Println(" ---------------------------------------------------------------------------------")
	fmt.Println("Please enter the message to be stored in the nodes:")
	//首先通过命令行获取用户输入
	stdReader := bufio.NewReader(os.Stdin)
	for {
		data, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
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
		//N0 is the primary node, and the request information is sent directly to N0 by default
		tcpDial(content, nodeTable["N0"])
	}
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
