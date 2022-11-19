#!/bin/sh

set -x 

BUILD_DIR=./build

cat /etc/os-release
cp /etc/os-release ${BUILD_DIR}/os-release


go build -o ${BUILD_DIR}/ ./...
