#!/bin/sh

set -x 

BUILD_DIR=./build

go build -o ${BUILD_DIR}/ ./...

cat /etc/os-release
cp /etc/os-release ${BUILD_DIR}/os-release

cat << EOF > ${BUILD_DIR}/info.json
{
   "BUILD_ID":   ${BUILD_ID}
   "GIT_COMMIT": ${GIT_COMMIT}
   "GIT_BRANCH": ${GIT_BRANCH}
   "GIT_URL":    ${GIT_URL}
}
EOF

ls -al ${BUILD_DIR}

