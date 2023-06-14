package fpbft

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strconv"
)

// If the 'Keys' directory does not exist in the current directory, create the directory,
// and generate RSA public and private keys for each node.
func genRsaKeys(numNodes int) {
	if !isExist("./Keys") {
		fmt.Println("the public and private key directory has not been generated yet, generating public and private keys...")
		err := os.Mkdir("Keys", 0755)
		if err != nil {
			log.Panic()
		}
		for i := 0; i <= numNodes; i++ {
			if !isExist("./Keys/N" + strconv.Itoa(i)) {
				err := os.Mkdir("./Keys/N"+strconv.Itoa(i), 0755)
				if err != nil {
					log.Panic()
				}
			}
			priv, pub := getKeyPair()
			privFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PIV"
			file, err := os.OpenFile(privFileName, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				log.Panic(err)
			}
			defer file.Close()
			file.Write(priv)

			pubFileName := "Keys/N" + strconv.Itoa(i) + "/N" + strconv.Itoa(i) + "_RSA_PUB"
			file2, err := os.OpenFile(pubFileName, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				log.Panic(err)
			}
			defer file2.Close()
			file2.Write(pub)
		}
		fmt.Println("RSA public and private keys have been generated for the nodes.")
	}
}

// Generate RSA public and private keys.
func getKeyPair() (prvkey, pubkey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}

// Determine whether the file or folder exists.
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}
