#!/bin/bash

set -x 

BUILD_DIR=./build

go mod tidy

go build -o ${BUILD_DIR}/ ./...
