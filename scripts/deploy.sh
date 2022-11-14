#!/bin/bash

set -x

NAME=players-tt-go

GROUPID=com.rsmaxwell.players
ARTIFACTID=${NAME}-x86_64-linux
VERSION=${BUILD_ID:-SNAPSHOT}
PACKAGING=zip

REPOSITORY=releases
REPOSITORYID=releases
URL=https://pluto.rsmaxwell.co.uk/archiva/repository/${REPOSITORY}


pwd
ls -al 


DIST_DIR=./dist
cd ${DIST_DIR}

ls -al 

ZIPFILE=$(ls ${NAME}*)

mvn --batch-mode deploy:deploy-file \
	-DgroupId=${GROUPID} \
	-DartifactId=${ARTIFACTID} \
	-Dversion=${VERSION} \
	-Dpackaging=${PACKAGING} \
	-Dfile=${ZIPFILE} \
	-DrepositoryId=${REPOSITORYID} \
	-Durl=${URL}
