#!/bin/bash

set -x 

BUILD_DIR=./build

go get tidy

go build -o ${BUILD_DIR}/ ./...
