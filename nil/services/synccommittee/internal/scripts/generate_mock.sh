#!/bin/bash

SRC_DIR="."
PKG_NAME=""

usage() {
    echo "Usage: $0 [-d <src directory>] [-p <package name>] <GoInterfaceName>"
    exit 1
}

while [ $# -gt 0 ]; do
    case "$1" in
    -d)
        if [ -z "$2" ]; then
            echo "Error: -d requires a directory argument"
            usage
        fi
        SRC_DIR="$2"
        shift 2
        ;;
    -p)
        if [ -z "$2" ]; then
            echo "Error: -p requires a package name argument"
            usage
        fi
        PKG_NAME="$2"
        shift 2
        ;;
    -*)
        echo "Unknown option: $1"
        usage
        ;;
    *)
        break
        ;;
    esac
done

# Check if the interface name is provided after options parsing
if [ "$#" -ne 1 ]; then
    usage
fi

INTERFACE_NAME="$1"

# Construct moq command based on whether package name is provided
MOQ_CMD="go run github.com/matryer/moq -rm -stub -with-resets"
if [ -n "$PKG_NAME" ]; then
    MOQ_CMD="$MOQ_CMD -pkg $PKG_NAME"
fi
MOQ_CMD="$MOQ_CMD $SRC_DIR $INTERFACE_NAME"

FILE_CONTENT="//go:build test
$($MOQ_CMD)"

snake_case() {
    echo "$1" | sed -E 's/([a-z])([A-Z])/\1_\2/g' | tr '[:upper:]' '[:lower:]'
}

echo "$FILE_CONTENT" >"$(snake_case "$INTERFACE_NAME")_generated_mock.go"
