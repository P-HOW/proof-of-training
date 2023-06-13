package pbft

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
)

// <REQUEST,o,t,c>
type Request struct {
	Message
	Timestamp int64
	//相当于clientID
	ClientAddr string
}

// <<PRE-PREPARE,v,n,d>,m>
type PrePrepare struct {
	RequestMessage Request
	Digest         string
	SequenceID     int
	Sign           []byte
}

// <PREPARE,v,n,d,i>
type Prepare struct {
	Digest     string
	SequenceID int
	NodeID     string
	Sign       []byte
}

// <COMMIT,v,n,D(m),i>
type Commit struct {
	Digest     string
	SequenceID int
	NodeID     string
	Sign       []byte
}

// <REPLY,v,t,c,i,r>
type Reply struct {
	MessageID int
	NodeID    string
	Result    bool
}

type Message struct {
	Content string
	ID      int
}

const prefixCMDLength = 12

type command string

const (
	cRequest    command = "request"
	cPrePrepare command = "preprepare"
	cPrepare    command = "prepare"
	cCommit     command = "commit"
)

// Join command and content in bytes.
// The default first twelve bytes are the command name.
func jointMessage(cmd command, content []byte) []byte {
	b := make([]byte, prefixCMDLength)
	for i, v := range []byte(cmd) {
		b[i] = v
	}
	joint := make([]byte, 0)
	joint = append(b, content...)
	return joint
}

// Split command and content in bytes.
// The default first twelve bytes are the command name.
func splitMessage(message []byte) (cmd string, content []byte) {
	cmdBytes := message[:prefixCMDLength]
	newCMDBytes := make([]byte, 0)
	for _, v := range cmdBytes {
		if v != byte(0) {
			newCMDBytes = append(newCMDBytes, v)
		}
	}
	cmd = string(newCMDBytes)
	content = message[prefixCMDLength:]
	return
}

// get message hash (ID)
func getDigest(request Request) string {
	b, err := json.Marshal(request)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	//hex encoding of the hash
	return hex.EncodeToString(hash[:])
}
