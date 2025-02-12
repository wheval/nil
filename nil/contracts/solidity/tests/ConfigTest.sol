// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.9;

import "../lib/Nil.sol";

contract ConfigTest is NilBase {

    function verifyExternal(uint256, bytes calldata) external pure returns (bool) {
        return true;
    }

    function testValidatorsEqual(Nil.ParamValidators memory inputValidators) public {
        Nil.ParamValidators memory realValidators = Nil.getValidators();
        require(inputValidators.validators.length == realValidators.validators.length, "Lengths are not equal");
        for (uint i = 0; i < inputValidators.validators.length; i++) {
        require(inputValidators.validators[i].list.length == realValidators.validators[i].list.length, "Lengths are not equal");
            for (uint j = 0; j < inputValidators.validators[i].list.length; j++) {
                Nil.ValidatorInfo memory input = inputValidators.validators[i].list[j];
                Nil.ValidatorInfo memory real = realValidators.validators[i].list[j];
                bytes32 a = keccak256(abi.encodePacked(input.PublicKey));
                bytes32 b = keccak256(abi.encodePacked(real.PublicKey));
                require(a == b, "Public keys are not equal");
                require(input.WithdrawalAddress == real.WithdrawalAddress, "Withdraw addresses are not equal");
            }
        }
    }

    function setValidators(Nil.ParamValidators memory validators) public {
        bytes memory data = abi.encode(validators);
        Nil.setConfigParam("curr_validators", data);
    }

    function testParamGasPriceEqual(Nil.ParamGasPrice memory param) public {
        Nil.ParamGasPrice memory realParam = Nil.getParamGasPrice();
        require(param.shards.length == realParam.shards.length, "Gas price shards length mismatch");
        for (uint i = 0; i < param.shards.length; i++) {
            require(param.shards[i] == realParam.shards[i], "Gas prices are not equal");
        }
    }

    function setParamGasPrice(Nil.ParamGasPrice memory param) public {
        bytes memory data = abi.encode(param);
        Nil.setConfigParam("gas_price", data);
    }

    function readParamAfterWrite() public {
        Nil.ParamGasPrice memory param = Nil.getParamGasPrice();
        param.shards[0] = 0x1234567890abcdef;
        bytes memory data = abi.encode(param);
        Nil.setConfigParam("gas_price", data);
        Nil.ParamGasPrice memory readParam = Nil.getParamGasPrice();
        require(readParam.shards[0] == 0x1234567890abcdef, "Gas price is not equal");
    }
}
