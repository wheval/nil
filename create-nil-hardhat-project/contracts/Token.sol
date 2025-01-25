// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "@nilfoundation/smart-contracts/contracts/NilTokenBase.sol";

contract Token is NilTokenBase {
    constructor(uint256 initialSupply) {
        // Mint the initial supply of tokens
        mintTokenInternal(initialSupply);
    }

    // Public function to call the parent internal function sendTokenInternal
    function transferToken(address to, TokenId tokenId, uint256 amount) public {
        sendTokenInternal(to, tokenId, amount);
    }
}
