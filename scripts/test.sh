#!/bin/sh

set -x

echo "testing: players-tt-api"

BUILD_DIR=./build

${BUILD_DIR}/players-tt-api --version
