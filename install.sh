#!/bin/bash

EXECUTABLE=kitlab
BIN_DIR=/usr/local/bin
BIN_EXECUTABLE="${BIN_DIR}/${EXECUTABLE}"

if test -f "${BIN_EXECUTABLE}"; then
    echo "Removing previous version"
    rm $BIN_EXECUTABLE
fi

ABS_PATH="$(cd "$(dirname "$1")"; pwd -P)/$(basename "$1")"

ln -s "${ABS_PATH}/kitlab" ${BIN_DIR}
echo "Successfully added 'kitlab' into ${BIN_DIR}"
