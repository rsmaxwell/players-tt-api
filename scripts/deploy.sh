#!/bin/bash

set -x

NAME=players-tt-api

GROUPID=com.rsmaxwell.players
ARTIFACTID=${NAME}-${1:=x86_64-linux}
VERSION=${BUILD_ID:-SNAPSHOT}
PACKAGING=zip

REPOSITORY=releases
REPOSITORYID=releases
URL=https://pluto.rsmaxwell.co.uk/archiva/repository/${REPOSITORY}

DIST_DIR=./dist
cd ${DIST_DIR}

ZIPFILE=$(ls ${NAME}*)

mvn --batch-mode deploy:deploy-file \
	-DgroupId=${GROUPID} \
	-DartifactId=${ARTIFACTID} \
	-Dversion=${VERSION} \
	-Dpackaging=${PACKAGING} \
	-Dfile=${ZIPFILE} \
	-DrepositoryId=${REPOSITORYID} \
	-Durl=${URL}
