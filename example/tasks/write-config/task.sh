#!/bin/bash -eu

ROOT=${PWD}

function main() {
  echo "writing to ${CONFIGURATION_FILENAME}..."
  echo "The date is $(date)" \
    > "${ROOT}/configuration/${CONFIGURATION_FILENAME}"
}

main
