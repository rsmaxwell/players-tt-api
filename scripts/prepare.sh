#!/bin/sh

set -x 

BRANCH=${1}
if [ -z "${BRANCH}" ]; then
    echo "Error: $0[${LINENO}]"
    echo "Missing BRANCH argument"
    exit 1
fi



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
		"GIT_BRANCH": "${BRANCH}",
		"GIT_URL": "$(git config --local remote.origin.url)"
	}
}
EOF

echo "git status"                        > ${BUILD_DIR}/info2.json
echo $(git status)                      >> ${BUILD_DIR}/info2.json
echo ""                                 >> ${BUILD_DIR}/info2.json
echo "git rev-parse --abbrev-ref HEAD)" >> ${BUILD_DIR}/info2.json
echo $(git rev-parse --abbrev-ref HEAD) >> ${BUILD_DIR}/info2.json
echo ""                                 >> ${BUILD_DIR}/info2.json


ls -al ${BUILD_DIR}/info.json
cat ${BUILD_DIR}/info.json
