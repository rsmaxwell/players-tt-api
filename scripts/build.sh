#!/bin/sh

set -x 

BUILD_DIR=./build

go build -o ${BUILD_DIR}/ ./...
