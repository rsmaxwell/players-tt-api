#!/bin/sh

set -x

NAME=players-tt-api

ZIPFILE=${NAME}.zip

BUILD_DIR="$(pwd)/build"
PACKAGE_DIR="$(pwd)/package"
DIST_DIR="$(pwd)/dist"

rm -rf ${PACKAGE_DIR} ${DIST_DIR}
mkdir -p ${PACKAGE_DIR} ${DIST_DIR}

cd ${PACKAGE_DIR}
cp -r ${BUILD_DIR}/* .

zip -r ${DIST_DIR}/${ZIPFILE} *

pwd
ls -al 
ls -al ${BUILD_DIR}
ls -al ${PACKAGE_DIR}