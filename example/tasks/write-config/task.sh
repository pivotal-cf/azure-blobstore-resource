#!/bin/bash -eu

ROOT=${PWD}

function main() {
  echo "writing to ${CONFIGURATION_FILENAME}..."
  echo "The date is $(data)" \
    > "${ROOT}/configuration/${CONFIGURATION_FILENAME}"
}

main
