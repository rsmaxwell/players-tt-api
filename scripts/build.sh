#!/bin/sh

set -x 

BUILD_DIR=./build

go build -o ${BUILD_DIR}/ ./...

cat /etc/os-release
cp /etc/os-release ${BUILD_DIR}/os-release

cat << EOF > ${BUILD_DIR}/info.json
{
   "version":${BUILD_ID}
}
EOF    

ls -al ${BUILD_DIR}

