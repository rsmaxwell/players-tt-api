#!/bin/sh

set -x 

BUILD_DIR=$(pwd)/build
INFO_DIR=${BUILD_DIR}/info

go build -o ${BUILD_DIR}/ ./...

pwd
ls -al 
ls -al ${BUILD_DIR}
ls -al ${INFO_DIR}

cp /etc/os-release ${INFO_DIR}/os-release

pwd
ls -al 
ls -al ${BUILD_DIR}
ls -al ${INFO_DIR}
