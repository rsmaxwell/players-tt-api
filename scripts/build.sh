#!/bin/sh

set -x 

BUILD_DIR=./build
INFO_DIR=./${BUILD_DIR}/info

go build -o ${BUILD_DIR}/ ./...

cp /etc/os-release ${INFO_DIR}/os-release

ls -al ${BUILD_DIR}
ls -al ${INFO_DIR}
