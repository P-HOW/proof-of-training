//Run this function only ONCE
var ethers = require('ethers');
var crypto = require('crypto');
var fs = require('fs');

var numKeys = 30;
var keys = [];

// Generate the keys
for (let i = 0; i < numKeys; i++) {
    var id = crypto.randomBytes(32).toString('hex');
    var privateKey = "0x"+id;

    var wallet = new ethers.Wallet(privateKey);
    keys.push({privateKey: privateKey, address: wallet.address});
}

// Write keys to a file
fs.writeFileSync('keys.json', JSON.stringify(keys, null, 2));

// Function to read keys back into an array
function getKeys() {
    var keysData = fs.readFileSync('keys.json');
    return JSON.parse(keysData);
}

// Use the function
var keysArray = getKeys();
console.log(keysArray);