// SPDX-License-Identifier: GPL-3.0

pragma solidity ^0.8.0;

import "@nilfoundation/smart-contracts/contracts/NilTokenBase.sol";

contract Token is NilTokenBase {
    bytes pubkey;

    constructor(string memory _tokenName, bytes memory _pubkey) payable {
        // Revert if the token name is an empty string
        require(bytes(_tokenName).length > 0, "Token name must not be empty");
        pubkey = _pubkey;
        tokenName = _tokenName;
    }
    receive() external payable {}

    function verifyExternal(uint256 hash, bytes calldata signature) external view returns (bool) {
        return Nil.validateSignature(pubkey, hash, signature);
    }
}
