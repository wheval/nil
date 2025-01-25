// SPDX-License-Identifier: GPL-3.0

pragma solidity ^0.8.9;

import "../lib/Nil.sol";

contract TransactionCheck is NilBase {
    function externalFunc() public onlyExternal {}

    function internalFunc() public onlyInternal {}

    // Fail: we call external method by sync call, which is considered as internal
    function callExternal(address addr) public onlyExternal {
        TransactionCheck(addr).externalFunc();
    }

    // Ok: we call internal method by sync call
    function callInternal(address addr) public onlyExternal {
        TransactionCheck(addr).internalFunc();
    }

    function verifyExternal(
        uint256,
        bytes calldata
    ) external pure returns (bool) {
        return true;
    }
}
