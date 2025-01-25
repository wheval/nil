// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Incrementer {
    uint256 public counter;

    constructor(uint256 start) {
        counter = start;
    }

    function increment() public {
        counter += 1;
    }

    function incrementExternal() public onlyExternal {
        counter += 1;
    }

    function getCounter() public view returns (uint256) {
        return counter;
    }

    function setCounter(uint256 _counter) public {
        counter = _counter;
    }

    function add(uint256 a, uint256 b) public pure returns (uint256) {
        return a + b;
    }

    receive() external payable {}

    function verifyExternal(
        uint256 hash,
        bytes calldata signature
    ) external view returns (bool) {
        return true;
    }

    modifier onlyExternal() {
        require(
            !isInternalTransaction(),
            "Trying to call external function with internal transaction"
        );
        _;
    }

    // isInternalTransaction returns true if the current transaction is internal.
    function isInternalTransaction() internal view returns (bool) {
        bytes memory data;
        (bool success, bytes memory returnData) = address(0xff).staticcall(
            data
        );
        require(success, "Precompiled contract call failed");
        require(
            returnData.length > 0,
            "'IS_INTERNAL_TRANSACTION' returns invalid data"
        );
        return abi.decode(returnData, (bool));
    }
}
