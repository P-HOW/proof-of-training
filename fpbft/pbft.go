package fpbft

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"sync"
)

type node struct {
	//Node ID
	nodeID string
	//Node Listening Address
	addr string
	//RSA private key
	rsaPrivKey []byte
	//RSA public key
	rsaPubKey []byte
}

type pbft struct {
	//node information
	node node
	//Each request increases the sequence number.
	sequenceID int
	//lock
	lock sync.Mutex
	//
	//Temporary message pool, the message digest corresponds to the message body.
	messagePool map[string]Request
	//
	//The number of prepares received (at least 2f are needed to be received and confirmed), corresponding according to the digest.
	prePareConfirmCount map[string]map[string]bool
	//
	//Stores the number of commits received (at least 2f+1 are needed to be received and confirmed), corresponding according to the digest.
	commitConfirmCount map[string]map[string]bool
	//
	//Has the broadcast for this message already been committed
	isCommitBordcast map[string]bool
	//
	//Has this message already been replied to the client?
	isReply map[string]bool

	nodeTable nodeTable

	nodeCount int

	// Local message pool (simulating the persistence layer), only after the confirmation of successful commit will the messages be stored in this pool.
	localMessagePool []Message

	// Temp prepare pool (simulating the unconfirmed layer), only after getting the prepare from view will this be moved forward.
	tempPreparePool []Prepare

	//Temp commit pool (simulating the unconfirmed layer), only after getting the map true
	tempCommitPool []Commit

	//Bandwidth of nodes, in Mbps
	bandwidth int

	//latency in milliseconds
	latency float64
}

func NewPBFT(nodeID, addr string, nodeTable nodeTable, nodeCount int, bandwidth float64, latency float64) *pbft {
	p := new(pbft)
	p.node.nodeID = nodeID
	p.node.addr = addr
	p.node.rsaPrivKey = p.getPivKey(nodeID) //Read from the generated private key file.
	p.node.rsaPubKey = p.getPubKey(nodeID)  //Read from the generated private key file.
	p.sequenceID = 0
	p.messagePool = make(map[string]Request)
	p.prePareConfirmCount = make(map[string]map[string]bool)
	p.commitConfirmCount = make(map[string]map[string]bool)
	p.isCommitBordcast = make(map[string]bool)
	p.isReply = make(map[string]bool)
	p.nodeTable = nodeTable
	p.nodeCount = nodeCount
	p.localMessagePool = []Message{}
	p.tempPreparePool = []Prepare{}
	p.tempCommitPool = []Commit{}
	p.bandwidth = int(bandwidth * 1024 * 1024 / 8)
	p.latency = latency
	return p
}

func (p *pbft) handleRequest(data []byte) {
	//Split the message and call different functions based on the message command.
	cmd, content := splitMessage(data)
	switch command(cmd) {
	case cRequest:
		p.handleClientRequest(content)
	case cPrePrepare:
		p.handlePrePrepare(content)
	case cPrepare:
		p.handlePrepare(content)
	case cCommit:
		p.handleCommit(content)
	}
}

// Handle requests coming from the client.
func (p *pbft) handleClientRequest(content []byte) {
	fmt.Println("The primary node has received a request from the client...")
	//Parsing the Request structure using JSON.
	r := new(Request)
	err := json.Unmarshal(content, r)
	if err != nil {
		log.Panic(err)
	}
	//add sequence number
	p.sequenceIDAdd()
	//fetch digest
	digest := getDigest(*r)
	fmt.Println("The request has been stored in the temporary message pool.")
	//Store in the temporary message pool.
	p.messagePool[digest] = *r
	//The primary node signs the message digest.
	digestByte, _ := hex.DecodeString(digest)
	signInfo := p.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
	//Assembled into PrePrepare, ready to be sent to follower nodes.
	pp := PrePrepare{*r, digest, p.sequenceID, signInfo}
	b, err := json.Marshal(pp)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Broadcasting PrePrepare to other nodes...")
	//Broadcast PrePrepare
	p.broadcast(cPrePrepare, b)
	fmt.Println("PrePrepare broadcast completed.")
}

// Process PrePrepare message
func (p *pbft) handlePrePrepare(content []byte) {
	//fmt.Println("This node has received the PrePrepare message sent by the primary node ...")
	//Parse out the PrePrepare structure using JSON
	pp := new(PrePrepare)
	err := json.Unmarshal(content, pp)
	if err != nil {
		log.Panic(err)
	}
	//To obtain the public key of the primary node for digital signature verification
	primaryNodePubKey := p.getPubKey("N0")
	digestByte, _ := hex.DecodeString(pp.Digest)
	if digest := getDigest(pp.RequestMessage); digest != pp.Digest {
		fmt.Println("The digest doesn't match, refuse to broadcast prepare")
	} else if p.sequenceID+1 != pp.SequenceID {
		fmt.Println("The message sequence number doesn't match, refuse to broadcast prepare")
	} else if !p.RsaVerySignWithSha256(digestByte, pp.Sign, primaryNodePubKey) {
		fmt.Println("The primary node signature verification failed! Refusing to broadcast prepare")
	} else {
		//Assigning the sequence number
		p.sequenceID = pp.SequenceID
		//Storing the information in the temporary message pool
		//fmt.Println("The message has been stored in the temporary node pool")
		p.messagePool[pp.Digest] = pp.RequestMessage
		//The node signs it with its private key
		// Handles the tempPreparePool and tempCommitPool and execute prepare or commit
		//it will be broadcasted by primary node so it will be executed only once
		p.handleTempPool()

		sign := p.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
		//Concatenate to form a Prepare message
		pre := Prepare{pp.Digest, pp.SequenceID, p.node.nodeID, sign}
		bPre, err := json.Marshal(pre)
		if err != nil {
			log.Panic(err)
		}

		//fmt.Println("broadcasting the Prepare message...")
		p.broadcast(cPrepare, bPre)
		//fmt.Println("Prepare broadcast is completed.")

	}
}

// Process the Prepare message
func (p *pbft) handlePrepare(content []byte) {
	//Parse out the Prepare structure using JSON
	pre := new(Prepare)
	err := json.Unmarshal(content, pre)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("The node has received Prepare from node %s ... \n", pre.NodeID)
	//To obtain the public key of the message source node for digital signature verification
	MessageNodePubKey := p.getPubKey(pre.NodeID)
	digestByte, _ := hex.DecodeString(pre.Digest)
	if _, ok := p.messagePool[pre.Digest]; !ok {
		p.tempPreparePool = append(p.tempPreparePool, *pre)
	} else if p.sequenceID != pre.SequenceID {
		fmt.Println("The message sequence number doesn't match. Refusing to execute commit broadcast")
	} else if !p.RsaVerySignWithSha256(digestByte, pre.Sign, MessageNodePubKey) {
		fmt.Println("The node signature verification failed! Refusing to execute commit broadcast")
	} else {
		p.prepareStageHandle(*pre, digestByte)
	}
}

// Processing the commit
func (p *pbft) handleCommit(content []byte) {
	//Parse out the Commit structure using JSON
	c := new(Commit)
	err := json.Unmarshal(content, c)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("The node has received Commit from node %s ... \n", c.NodeID)
	//To obtain the public key of the message source node for digital signature verification
	MessageNodePubKey := p.getPubKey(c.NodeID)
	digestByte, _ := hex.DecodeString(c.Digest)

	if _, ok := p.prePareConfirmCount[c.Digest]; !ok {
		p.tempCommitPool = append(p.tempCommitPool, *c)
	} else if p.sequenceID != c.SequenceID {
		fmt.Println("The message sequence number doesn't match. Refusing to persist the information to the local message pool")
	} else if !p.RsaVerySignWithSha256(digestByte, c.Sign, MessageNodePubKey) {
		fmt.Println("The node signature verification failed! Refusing to persist the information to the local message pool")
	} else {
		p.setCommitConfirmMap(c.Digest, c.NodeID, true)
		count := 0
		for range p.commitConfirmCount[c.Digest] {
			count++
		}
		//If a node has received at least 2f+1 commit messages (including itself), and the node has not replied before,
		//and a commit broadcast has been performed, then the information is submitted to the local message pool,
		//and a successful flag is replied to the client!
		p.lock.Lock()
		if count >= p.nodeCount/3*2 && !p.isReply[c.Digest] && p.isCommitBordcast[c.Digest] {
			//fmt.Println("This node has received at least 2f + 1 Commit messages (including the local node) from other nodes ...")
			//The message information is being submitted to the local message pool!
			p.localMessagePool = append(p.localMessagePool, p.messagePool[c.Digest].Message)
			info := p.node.nodeID + "node has put msgid:" + strconv.Itoa(p.messagePool[c.Digest].ID) + "into the local message pool,message content：" + p.messagePool[c.Digest].Content
			//fmt.Println(info)
			//fmt.Println("Replying to client ...")
			tcpDial([]byte(info), p.messagePool[c.Digest].ClientAddr, p.bandwidth, p.latency)
			p.isReply[c.Digest] = true
			//fmt.Println("replying done!")
		}
		p.lock.Unlock()
	}
}

// Add sequenceID
func (p *pbft) sequenceIDAdd() {
	p.lock.Lock()
	p.sequenceID++
	p.lock.Unlock()
}

// Broadcasting to other nodes except itself
func (p *pbft) broadcast(cmd command, content []byte) {
	message := jointMessage(cmd, content)
	for i := range p.nodeTable {
		if i == p.node.nodeID {
			continue
		}

		go tcpDial(message, p.nodeTable[i], p.bandwidth, p.latency)
	}
}

// Allocating assignment for multiple mappings
func (p *pbft) setPrePareConfirmMap(val, val2 string, b bool) {
	if _, ok := p.prePareConfirmCount[val]; !ok {
		p.prePareConfirmCount[val] = make(map[string]bool)
	}
	p.prePareConfirmCount[val][val2] = b
}

// Allocating assignment for multiple mappings
func (p *pbft) setCommitConfirmMap(val, val2 string, b bool) {
	if _, ok := p.commitConfirmCount[val]; !ok {
		p.commitConfirmCount[val] = make(map[string]bool)
	}
	p.commitConfirmCount[val][val2] = b
}

// Pass the node number to obtain the corresponding public key
func (p *pbft) getPubKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PUB")
	if err != nil {
		log.Panic(err)
	}
	return key
}

// Pass the node number and obtain the corresponding private key
func (p *pbft) getPivKey(nodeID string) []byte {
	key, err := ioutil.ReadFile("Keys/" + nodeID + "/" + nodeID + "_RSA_PIV")
	if err != nil {
		log.Panic(err)
	}
	return key
}

// Digital signature
func (p *pbft) RsaSignWithSha256(data []byte, keyBytes []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

// Verify signature
func (p *pbft) RsaVerySignWithSha256(data, signData, keyBytes []byte) bool {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		panic(err)
	}
	return true
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

func (p *pbft) prepareStageHandle(pre Prepare, digestByte []byte) {

	p.setPrePareConfirmMap(pre.Digest, pre.NodeID, true)

	count := p.getPrepareCount(pre)

	specifiedCount := p.getSpecifiedPrepareCount()

	//To obtain the public key of the message source node for digital signature verification
	if count >= specifiedCount && !p.isCommitBordcast[pre.Digest] {

		p.finalizePrepare(digestByte, pre)

	}

}

func (p *pbft) getSpecifiedPrepareCount() int {
	//Because the primary node does not send Prepare, it does not include itself
	specifiedCount := 0
	if p.node.nodeID == "N0" {
		specifiedCount = p.nodeCount / 3 * 2
	} else {
		specifiedCount = (p.nodeCount / 3 * 2) - 1
	}
	return specifiedCount
}

func (p *pbft) getPrepareCount(pre Prepare) int {
	count := 0
	for range p.prePareConfirmCount[pre.Digest] {
		count++
	}
	return count
}

func (p *pbft) finalizePrepare(digestByte []byte, pre Prepare) {
	//If a node has received at least 2f Prepare messages (including itself) and
	//has not yet performed a commit broadcast, it will proceed with a commit broadcast
	p.lock.Lock()
	//fmt.Println("This node has received at least 2f Prepare messages (including the local node) from other nodes ...")
	//The node signs it with its private key
	sign := p.RsaSignWithSha256(digestByte, p.node.rsaPrivKey)
	c := Commit{pre.Digest, pre.SequenceID, p.node.nodeID, sign}
	bc, err := json.Marshal(c)
	if err != nil {
		log.Panic(err)
	}
	//Broadcasting the commit message
	//fmt.Println("broadcasting the commit message...")
	p.broadcast(cCommit, bc)
	p.isCommitBordcast[pre.Digest] = true
	//fmt.Println("commit broadcast is completed")
	p.lock.Unlock()
}

func (p *pbft) commitStageHandle(c Commit) {

	p.setCommitConfirmMap(c.Digest, c.NodeID, true)

	count := p.getCommitCount(c)

	specifiedCount := p.getSpecifiedCommitCount()

	//If a node has received at least 2f+1 commit messages (including itself), and the node has not replied before,
	//and a commit broadcast has been performed, then the information is submitted to the local message pool,
	//and a successful flag is replied to the client!

	if count >= specifiedCount && !p.isReply[c.Digest] && p.isCommitBordcast[c.Digest] {
		p.finalizeCommit(c)
	}

}

func (p *pbft) getCommitCount(c Commit) int {
	count := 0
	for range p.commitConfirmCount[c.Digest] {
		count++
	}
	return count
}

func (p *pbft) getSpecifiedCommitCount() int {
	return p.nodeCount / 3 * 2
}

func (p *pbft) finalizeCommit(c Commit) {
	p.lock.Lock()
	//fmt.Println("This node has received at least 2f + 1 Commit messages (including the local node) from other nodes ...")
	//The message information is being submitted to the local message pool!
	p.localMessagePool = append(p.localMessagePool, p.messagePool[c.Digest].Message)
	info := p.node.nodeID + "node has put msgid:" + strconv.Itoa(p.messagePool[c.Digest].ID) + "into the local message pool,message content：" + p.messagePool[c.Digest].Content
	//fmt.Println(info)
	//fmt.Println("Replying to client ...")
	tcpDial([]byte(info), p.messagePool[c.Digest].ClientAddr, p.bandwidth, p.latency)
	p.isReply[c.Digest] = true
	//fmt.Println("replying done!")
	p.lock.Unlock()
}

func (p *pbft) handleTempPool() {
	p.lock.Lock()
	for _, prepare := range p.tempPreparePool {
		content, _ := json.Marshal(prepare)
		p.handlePrepare(content)
	}

	p.tempPreparePool = []Prepare{}

	for _, commit := range p.tempCommitPool {
		content, _ := json.Marshal(commit)
		p.handleCommit(content)
	}

	p.tempCommitPool = []Commit{}

	p.lock.Unlock()
}
