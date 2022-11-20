#!/bin/sh

set -x 
 
VERSION="0.0.$((${BUILD_ID}))"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
find . -name "version.go" | while read versionfile; do

    echo "Replacing tags in ${versionfile}"

    sed -i "s@<VERSION>@${VERSION}@g"            ${versionfile}
    sed -i "s@<BUILD_ID>@${BUILD_ID}@g"          ${versionfile}
    sed -i "s@<BUILD_DATE>@${TIMESTAMP}@g"       ${versionfile}
    sed -i "s@<GIT_COMMIT>@${GIT_COMMIT}@g"      ${versionfile}
    sed -i "s@<GIT_BRANCH>@${GIT_BRANCH}@g"      ${versionfile}
    sed -i "s@<GIT_URL>@${GIT_URL}@g"            ${versionfile}
done

BUILD_DIR=./build
mkdir -p ${BUILD_DIR}

cat << EOF > ${BUILD_DIR}/info.json
{
	"VERSION": "${VERSION}",
	"BUILD_ID": ${BUILD_ID},
	"TIMESTAMP": "${TIMESTAMP}",
	"pipeline": {
		"GIT_COMMIT": "${GIT_COMMIT}",
		"GIT_BRANCH": "${GIT_BRANCH}",
		"GIT_URL": "${GIT_URL}"
	},
	"project": {
		"GIT_COMMIT": "$(git rev-parse HEAD)",
		"GIT_BRANCH": "$(git rev-parse --abbrev-ref HEAD)",
		"GIT_URL": "$(git config --local remote.origin.url)"
	}
}
EOF


git rev-parse HEAD
git rev-parse --abbrev-ref HEAD
git config --local remote.origin.url


ls -al ${BUILD_DIR}/info.json
cat ${BUILD_DIR}/info.json
