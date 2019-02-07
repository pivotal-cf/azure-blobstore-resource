#!/bin/bash -eu

ROOT=${PWD}

function main() {
  echo "printing ${CONFIGURATION_FILENAME}..."
  cat ${ROOT}/configuration/${CONFIGURATION_FILENAME}
}

main
