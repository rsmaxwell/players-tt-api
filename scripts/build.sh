#!/bin/bash

set -x 

BUILD_DIR=./build

go get -u github.com/gorilla/mux

go build -o ${BUILD_DIR}/ ./...
