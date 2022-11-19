#!/bin/sh

set -x 

BUILD_DIR=./build

sh('cat /etc/os-release')
sh('cp /etc/os-release ${BUILD_DIR}/os-release')


go build -o ${BUILD_DIR}/ ./...
