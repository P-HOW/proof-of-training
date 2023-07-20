# Proof-of-Training (POT)

![Go](https://img.shields.io/badge/Go-v1.20+-blue.svg)
[![Build Status](https://travis-ci.org/anfederico/clairvoyant.svg?branch=master)](https://travis-ci.org/anfederico/clairvoyant)
![Dependencies](https://img.shields.io/badge/dependencies-up%20to%20date-brightgreen.svg)
[![GitHub Issues](https://img.shields.io/github/issues/P-HOW/proof-of-training.svg)](https://github.com/P-HOW/proof-of-training/issues)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

:star: Star us on GitHub — it motivates us a lot!

Golang execution layer implementation of the decentralized training network using proof of training (POT).

![DTN](https://github.com/P-HOW/proof-of-training/blob/master/img/dtn.jpg?raw=true)

## Table Of Content
- [Introduction](#introduction)
- [Layer-1(L1) Implementation and Tests](#layer-1-implementation-and-tests)
    - [Practical Byzantine Fault Tolerance (PBFT)](#pbft-package)
    - [(Recommended) Full Practical Byzantine Fault Tolerance (FPBFT)](#fpbft-package)
- [Layer-2(L2) Implementation and Tests](#layer-2-implementation-and-tests)
    - [Environment Setup](#setup)
    - [Compile and Deploy in the Truffle Framework](#compile-and-deploy)
    - [Contract Testing](#test)
- [License](#license)
- [Links](#links)
## Introduction
This repo provides a miniature kernel realization of various components of the Decentralized Training Network (DTN) based on Proof-of-Training (POT). It primarily focuses on several fundamental aspects of both our Layer-1 (L1) and Layer-2 (L2) implementations.

- **Global Ledger Maintenance and Synchronization**
- Multisignature Coordination

## Layer-1 Implementation and Tests


Our L1 implementation uses the [Practical Byzantine Fault Tolerance (PBFT)](https://pmg.csail.mit.edu/papers/osdi99.pdf) protocol, which ensures consensus among nodes in a distributed network, even in the presence of malicious nodes or if certain nodes fail.


### PBFT Package

The PBFT package ensures the uniformity of our global ledger by facilitating consensus among nodes on the set of transactions to be added. These approved transactions are synced to the global transaction pool, from which they are used to update the global ledger, ensuring data consistency across the network.

#### Functionality Implemented:
> PBFT formula: n >= 3f + 1 where n is the total number of nodes in the entire network, and f is the maximum number of malicious or faulty nodes allowed.

The data from client input to receiving replies from the nodes is divided into 5 steps:

1. The client sends request information to the primary node.
2. After the primary node N0 receives the request from the client, it extracts the main information from the request data and sends a preprepare to the other nodes.
3. The secondary nodes receive the preprepare from the primary node, firstly using the primary node's public key for signature authentication, then hash the message (message digest, to reduce the size of the information transmitted in the network), and broadcast a prepare to other nodes.
4. When a node receives 2f prepare information (including itself), and all signature verifications pass, it can proceed to the commit step, broadcasting a commit to all other nodes in the network.
5. When a node receives 2f+1 commit information (including itself), and all signature verifications pass, it can store the message locally, and return a reply message to the client.
   
![DTN](https://github.com/P-HOW/proof-of-training/blob/master/img/PBFTflow.png?raw=true)

To spawn a network applying PBFT with `numNodes` nodes, and to synchronize 
the transaction pool with transactions encoded in `data`, 
it is recommended to use the following function. This will help 
analyze the time needed for `data` synchronization.

Field | Data Types | Recommended Value
----  |------------| ----------
numNodes  | int        | 5-100
data  | string     | -
clientAddr  | string     | "127.0.0.1:8888"

```go
synctime := genPBFTSynchronize(numNodes, data, clientAddr)
```
> Note: The current implementations in the pbft package may contain 
> potential race conditions, potentially leading to non-terminating 
> execution. It is essential to implement a timing mechanism when using 
> this function for efficient operation.

#### pbft_test.go
```go
package pbft

import (
  "strconv"
  "testing"
)

func TestAddAndGetMessage(t *testing.T) {
  var clientAddr = "127.0.0.1:8888"
  var data = "transactions to be synchronized"
  var numNodes = 8
  sync_time := genPBFTSynchronize(numNodes, data, clientAddr)
  s := strconv.FormatFloat(sync_time, 'f', -1, 64)
  println("It takes " + s + " seconds to synchronize the transactions to the global ledger")
}
```
#### output
```text
=== RUN   TestAddAndGetMessage
the public and private key directory has not been generated yet, generating public and private keys...
RSA public and private keys have been generated for the nodes.
initiating client...
{"Content":"transactions to be synchronized","ID":2818938398,"Timestamp":1686148966474242384,"ClientAddr":"127.0.0.1:8888"}
The primary node has received a request from the client...
The request has been stored in the temporary message pool.
Broadcasting PrePrepare to other nodes...
PrePrepare broadcast completed.
It takes 0.032963246 seconds to synchronize the transactions to the global ledger
--- PASS: TestAddAndGetMessage (0.12s)
PASS
```

### FPBFT Package
FPBFT is a comprehensive simulation of the PBFT (Practical Byzantine Fault Tolerance) algorithm, designed to emulate real-world network conditions in distributed consensus scenarios. Unlike other PBFT implementations, FPBFT emphasizes the importance of varying network conditions, specifically network latency and bandwidth limit, which significantly impact the performance and fault tolerance of a distributed system. In real-world scenarios, nodes in a distributed network are dispersed across various geographical regions, each experiencing different network conditions. FPBFT integrates these parameters into the PBFT network generation, thereby setting itself apart as a full implementation of the PBFT algorithm.
#### Functionality Implemented:
> In addition to PBFT package, the FPBFT package emulated a real network by adding bandwidthLimit and latency 
> as input parameters when generating the network.

Field | Data Types | Sample Value
----  |------------| ----------
numNodes  | int        | 30-1000
data  | string     | "transactions"
clientAddr  | string     | "127.0.0.1:8888"
bandwidthLimit  | float64    | 20 (Mbps)
latency  | float64    | 350 (ms)

```go
synctime := genPBFTSynchronize(numNodes int, data string, clientAddr string, bandwidthLimit float64, latency float64)
```
> Note: by setting `bandwidthLimit` and `latency` to 0, 
> the function becomes PBFT as a special case.

#### fpbft_test.go
```go
package fpbft

import (
  "strconv"
  "testing"
)

func TestAddAndGetMessage(t *testing.T) {
  var clientAddr = "127.0.0.1:8888"
  var data = "transactions to be synchronized"
  var numNodes = 10
  sync_time := genPBFTSynchronize(numNodes, data, clientAddr, 0.01, 300)
  s := strconv.FormatFloat(sync_time, 'f', -1, 64)
  println("It takes " + s + " seconds to synchronize the transactions to the global ledger")
}

```
#### output
```text
=== RUN   TestAddAndGetMessage
initiating client...
The primary node has received a request from the client...
The request has been stored in the temporary message pool.
Broadcasting PrePrepare to other nodes...
PrePrepare broadcast completed.
It takes 2.264332033 seconds to synchronize the transactions to the global ledger
--- PASS: TestAddAndGetMessage (2.26s)
PASS
```
Sample experiments were conducted on a 64-bit Ubuntu 22.04.2 LTS system powered by the 12th Generation Intel® Core™ i7-12700T 
processor with 20 cores, and equipped with 32GB memory. With the following performance matrix:

| Message Size | Network Size | Slow (0.1 Mbps) | Medium (30 Mbps) | Fast (125 Mbps) |
|--------------|--------------|-----------------|------------------|-----------------|
| 100 transactions | Small (10 nodes) | 8.609 s | 1.494 s | 1.497 s |
| 100 transactions | Medium (30 nodes) | 8.707 s | 1.685 s | 1.755 s |
| 1000 transactions | Medium (30 nodes) | 73.536 s | 1.682 s | 1.833 s |
| 100 transactions | Large (50 nodes) | 8.697 s | 1.842 s | 1.752 s |
| 200 transactions | Large (50 nodes) | 15.984 s | 1.908 s | 1.893 s |
| 5000 transactions | Large (50 nodes) | 37.532 s | 1.767 s | 1.678 s |
| 10000 transactions | Large (50 nodes) | - | 7.215 s | 2.074 s |





## Layer-2 Implementation and Tests
This section provides a concrete implementation of a Layer-2 in the DTN. It leverages Ethereum 
smart contracts, deployed on the Binance Smart Chain (BSC) Testnet, to handle token balances 
and transactions more efficiently. In this guide, we go through how to set up the development 
environment and run the tests, ensuring that the Layer-2 solution functions as expected. 
We'll be using Truffle, a popular Ethereum development environment, to compile, deploy, 
and test our smart contracts on the BSC Testnet.
### Setup
Before you begin, make sure you have installed the latest version of [Node.js and npm](https://nodejs.org/en/download/). We put our implementations in the `./crypto/` directory. Change into the directory and install [truffle](https://trufflesuite.com/docs/truffle/quickstart/):

```bash
npm install -g truffle
```

If you want to run the tests by deploying the smart contracts to BSC testnet on your own, you need to 
generate the keys for testing by running 
```bash
node wallet/generate.js 
```
which helps you generate 30 random keys with private keys and addresses stored in `keys.json`. 

### Compile and Deploy
To compile the contract, run:
```bash
truffle compile
```
The `truffle-config.js` is written to
migrate the contracts to BSC testnet so you might consider getting some test BNB(TBNB) by getting from the [BNB faucet](https://testnet.binance.org/faucet-smart/) with the addresses you just generated.
Then, replace the `privateKeys` in the config file with the one you have TBNB in. Once set, run:
```bash
truffle migrate --network testnet
```
#### output
```text
> Compiled successfully using:
   - solc: 0.8.13+commit.abaa5c0e.Emscripten.clang


Starting migrations...
======================
> Network name:    'testnet'
> Network id:      97
> Block gas limit: 50000000 (0x2faf080)


1_deploy_contracts.js
=====================

   Deploying 'MultiSigContract'
   ----------------------------
   > transaction hash:    0xd0598d6c326e10c0fc64f1854de44cbbce7810b1307750ae0a48283a2e707d79
   > Blocks: 3            Seconds: 11
   > contract address:    0xDf033A1959006CD99c2549d5F6B427978b1cE2e8
   > block number:        30812543
   > block timestamp:     1687132610
   > account:             0xDF86931B3Bd2a3A65b7313cE8cBf7B63a6B203ef
   > balance:             0.39154595
   > gas used:            2169081 (0x2118f9)
   > gas price:           50 gwei
   > value sent:          0 ETH
   > total cost:          0.10845405 ETH

   > Saving artifacts
   -------------------------------------
   > Total cost:          0.10845405 ETH

Summary
=======
> Total deployments:   1
> Final cost:          0.10845405 ETH
```
and you will be able to check the deployment details from [bscscan](https://testnet.bscscan.com/) by searching the transaction hash:

![Install Aimeos TYPO3 extension](https://github.com/P-HOW/proof-of-training/blob/master/img/testbsc.png?raw=true)

### Test
To run the test, first distribute TBNB to other multi-sig wallets.
```bash
node transfer.js
```
Then test functions in the delpoyed contract by running:
```bash
truffle test --network testnet
```
Three contract functions are involved in the rewards distribution process for `_destination(address)`, which successfully generated the best AI model for a certain order:

![Reward Update Image](https://github.com/P-HOW/proof-of-training/blob/master/img/rewardUpdate.png?raw=true)

| Function Name | Input Parameters | Gas Cost | Number of Executions |
| --- | --- | --- |----------------------|
| proposeTransaction | _destination (address), _value (uint) | 86,875 | 1                    |
| confirmTransaction | _txIndex (uint) |  45,371 | `numSignaturesRequired`                  |
| executeTransaction | _txIndex (uint) | 161,888 | 1                    |

The tests provided summarizes the gas costs and execution counts for the key functions within 
the MultiSigContract deployed on the BSC testnet. These metrics can help in estimating 
the real costs of operations on BSC or Ethereum mainnet, as the actual cost can be 
by multiplying the gas cost with the current gas price. 
## License

The decentralized training network (DTN) as an implementation of proof of training (POT) is licensed under the terms of MIT
license and is available for free.

## Links

* [Paper](https://arxiv.org/pdf/2307.07066.pdf)
* [PBFT Reference Documentation](https://pmg.csail.mit.edu/papers/osdi99.pdf)
* [Testnet Contract](https://testnet.bscscan.com/address/0xdf033a1959006cd99c2549d5f6b427978b1ce2e8)
* [Issue tracker](https://github.com/P-HOW/proof-of-training/issues)
* [Author Info](https://lipeihao.com/)
