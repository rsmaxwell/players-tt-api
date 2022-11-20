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

cat << EOF > ${BUILD_DIR]/info.json
{
	"VERSION": ${VERSION}
	"BUILD_ID": ${BUILD_ID}
	"TIMESTAMP": ${TIMESTAMP}
	"GIT_COMMIT": ${GIT_COMMIT}
	"GIT_BRANCH": ${GIT_BRANCH}
	"GIT_URL": ${GIT_URL}
}
EOF