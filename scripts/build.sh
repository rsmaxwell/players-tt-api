#!/bin/sh

set -x 

BUILD_DIR=./build

go build -o ${BUILD_DIR}/ ./...

cat /etc/os-release
cp /etc/os-release ${BUILD_DIR}/os-release

ls -al ${BUILD_DIR}

