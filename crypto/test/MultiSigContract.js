const MultiSigWallet = artifacts.require("MultiSigContract");
const fs = require('fs');
const Web3 = require('web3');

// parse keys from keys.json
const keys = JSON.parse(fs.readFileSync('keys.json', 'utf8'));

// Assign the owners
const owner1 = keys[0].privateKey;
const owner2 = keys[1].privateKey;
const recipient = keys[2].privateKey;

// Connect to BSC Testnet
const web3 = new Web3('https://data-seed-prebsc-1-s3.binance.org:8545/');

// Your contract address
const contractAddress = '0xDf033A1959006CD99c2549d5F6B427978b1cE2e8';

contract('MultiSigWallet', () => {
  let multisigWallet;
  const requiredConfirmations = 2;
  const transferAmount = web3.utils.toWei('0.1', 'ether');
  console.log(transferAmount)
  beforeEach(async () => {
    multisigWallet = await MultiSigWallet.at(contractAddress);
    console.log(multisigWallet)
  });

  it('should propose transaction', async () => {
    const proposeTx = await multisigWallet.submitTransaction(recipient, transferAmount, '0x', { from: owner1 });
    const gasUsed = proposeTx.receipt.gasUsed;
    console.log(`Gas used for proposing transaction: ${gasUsed}`);
    assert(proposeTx.receipt.status, true);
  });

  it('should confirm transaction', async () => {
    const proposeTx = await multisigWallet.submitTransaction(recipient, transferAmount, '0x', { from: owner1 });
    const transactionId = proposeTx.logs[0].args.transactionId.toNumber();
    const confirmTx = await multisigWallet.confirmTransaction(transactionId, { from: owner2 });
    const gasUsed = confirmTx.receipt.gasUsed;
    console.log(`Gas used for confirming transaction: ${gasUsed}`);
    assert(confirmTx.receipt.status, true);
  });

  it('should execute transaction', async () => {
    const proposeTx = await multisigWallet.submitTransaction(recipient, transferAmount, '0x', { from: owner1 });
    const transactionId = proposeTx.logs[0].args.transactionId.toNumber();
    await multisigWallet.confirmTransaction(transactionId, { from: owner2 });
    const executeTx = await multisigWallet.executeTransaction(transactionId, { from: owner1 });
    const gasUsed = executeTx.receipt.gasUsed;
    console.log(`Gas used for executing transaction: ${gasUsed}`);
    assert(executeTx.receipt.status, true);
  });
});
