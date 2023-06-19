const Web3 = require('web3');
const fs = require('fs');

// Parse keys from keys.json
const keys = JSON.parse(fs.readFileSync('keys.json', 'utf8'));

// Get owner private keys and addresses
const owner1PrivateKey = keys[0].privateKey;
const owner2Address = keys[1].address;
const owner3Address = keys[2].address;

// Connect to BSC Testnet
const web3 = new Web3('https://data-seed-prebsc-1-s1.binance.org:8545');

// Add private key to web3 (remove '0x' prefix)
const account = web3.eth.accounts.privateKeyToAccount(owner1PrivateKey.replace('0x', ''));

// Add account to wallet
web3.eth.accounts.wallet.add(account);

// Set default account
web3.eth.defaultAccount = account.address;

// Define transfer amounts
const amountInTBNB = web3.utils.toWei('0.1', 'ether');  // Adjust this value as needed

async function transferTBNB() {
    // Transfer TBNB to owner2
    const tx1 = {
        to: owner2Address,
        value: amountInTBNB,
        gas: 21000
    };
    const receipt1 = await web3.eth.sendTransaction(tx1);
    console.log(`Transaction hash (owner1 to owner2): ${receipt1.transactionHash}`);

    // Transfer TBNB to owner3
    const tx2 = {
        to: owner3Address,
        value: amountInTBNB,
        gas: 21000
    };
    const receipt2 = await web3.eth.sendTransaction(tx2);
    console.log(`Transaction hash (owner1 to owner3): ${receipt2.transactionHash}`);
}

transferTBNB().catch(console.error);
