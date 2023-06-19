const fs = require('fs');
const MultiSigContract = artifacts.require("MultiSigContract");

module.exports = function(deployer) {
  // Read the keys from the file
  const keys = JSON.parse(fs.readFileSync('keys.json', 'utf8'));

  // Convert the keys into an array of addresses
  const owners = keys.map(key => key.address);
  const numSignaturesRequired = 3;

  deployer.deploy(MultiSigContract, numSignaturesRequired, owners);
};
