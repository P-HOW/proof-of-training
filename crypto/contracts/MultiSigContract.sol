pragma solidity ^0.8.4;

// Interface for ERC20 token contract

contract MultiSigContract {
	uint256 public numSignaturesRequired;
	address[] public owners;
	mapping (address => bool) public isOwner;

	//IERC20 public token; // ERC20 token contract address

	struct Transaction {
		address destination;
		uint value;
		bool executed;
	}
	Transaction[] public transactions;
	mapping (uint => mapping (address => bool)) public confirmations;

	// The balance that each address is allowed to withdraw
	mapping (address => uint) public rewards;

	constructor(uint256 _numSignaturesRequired, address[] memory _owners) {
		require(_owners.length >= _numSignaturesRequired && _numSignaturesRequired > 0,
			"Invalid number of owners and signatures required.");
		numSignaturesRequired = _numSignaturesRequired;

		for (uint i = 0; i < _owners.length; i++) {
			isOwner[_owners[i]] = true;
			owners.push(_owners[i]);
		}
	}

	function proposeTransaction(address _destination, uint _value) public onlyOwners {
		transactions.push(Transaction({
			destination: _destination,
			value: _value,
			executed: false
		}));
	}

	function confirmTransaction(uint _txIndex) public onlyOwners {
		require(!transactions[_txIndex].executed, "Transaction already executed.");
		confirmations[_txIndex][msg.sender] = true;
	}

	function executeTransaction(uint _txIndex) public onlyOwners {
		require(!transactions[_txIndex].executed, "Transaction already executed.");
		if (isConfirmed(_txIndex)) {
			Transaction storage tx = transactions[_txIndex];
			tx.executed = true;
			rewards[tx.destination] += tx.value;
		}
	}

	function isConfirmed(uint _txIndex) internal view returns (bool) {
		uint count = 0;
		for (uint i = 0; i < owners.length; i++) {
			if (confirmations[_txIndex][owners[i]]) {
				count += 1;
			}
			if (count == numSignaturesRequired) {
				return true;
			}
		}
		return false;
	}

	modifier onlyOwners() {
		require(isOwner[msg.sender], "Only owners can execute this action");
		_;
	}
}
