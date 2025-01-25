#!/bin/bash

set -e

SCRIPT_DIR=$(dirname $0)
NIL=${1:-nil}

$NIL cometa register --compile-input $SCRIPT_DIR/solidity/compile-faucet.json --address 0x0001111111111111111111111111111111111110
$NIL cometa register --compile-input $SCRIPT_DIR/solidity/compile-smart-account.json --address 0x0001111111111111111111111111111111111111
