#!/usr/bin/env bash

set -e

trap_with_arg() {
    local func="$1"
    shift
    for sig in "$@"; do
        trap "$func $sig" "$sig"
    done
}

stop() {
    trap - SIGINT EXIT
    printf '\n%s\n' "received $1, killing child processes"
    kill -s SIGINT $(jobs -pr)
}

trap_with_arg 'stop' EXIT SIGINT SIGTERM SIGHUP

# Clean up after previous runs
rm -f config.ini
rm -rf test.db

# Start nild in background (will be auto-killed on exit)
nild run --http-port 8529 --collator-tick-ms=100 >nild.log 2>&1 &
NILD_PID=$!
sleep 2

export NIL_RPC_ENDPOINT=http://127.0.0.1:8529
nil -c config.ini config set rpc_endpoint "$NIL_RPC_ENDPOINT"
export PRIVATE_KEY=$(nil -c config.ini keygen new -q)
export SMART_ACCOUNT_ADDR=$(nil -c config.ini smart-account new -q)

echo "Rpc endpoint: $NIL_RPC_ENDPOINT"
echo "Private key: $PRIVATE_KEY"
echo "Smart account addr: $SMART_ACCOUNT_ADDR"

# Update to reflect the new directory structure
# Move to the directory where the script is located
cd $(dirname "$0")

set +e
if CI=true pnpm exec hardhat test --network nil test/*.ts; then
    exit 0
else
    STATUS=$?
    kill $NILD_PID
    cat nild.log
    exit $STATUS
fi
