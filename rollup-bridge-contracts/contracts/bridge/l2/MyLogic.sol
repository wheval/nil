// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { PausableUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import { EnumerableSet } from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import { AccessControlEnumerableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/extensions/AccessControlEnumerableUpgradeable.sol";
import { IERC165 } from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import { NilAccessControlUpgradeable } from "../../NilAccessControlUpgradeable.sol";
import { NilConstants } from "../../common/libraries/NilConstants.sol";
import { ErrorInvalidMessageType } from "../../common/NilErrorConstants.sol";
import { AddressChecker } from "../../common/libraries/AddressChecker.sol";
import { StorageUtils } from "../../common/libraries/StorageUtils.sol";
import { IL2BridgeMessenger } from "./interfaces/IL2BridgeMessenger.sol";
import { IBridgeMessenger } from "../interfaces/IBridgeMessenger.sol";
import { IL2Bridge } from "./interfaces/IL2Bridge.sol";
import { IBridge } from "../interfaces/IBridge.sol";
import { INilMessageTree } from "../../interfaces/INilMessageTree.sol";

contract MyLogic is OwnableUpgradeable, PausableUpgradeable, NilAccessControlUpgradeable, ReentrancyGuardUpgradeable {
  uint256 public value;

  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  function initialize(uint256 _value) public initializer {
    value = _value;
  }

  function setValue(uint256 _value) public {
    value = _value;
  }
}
