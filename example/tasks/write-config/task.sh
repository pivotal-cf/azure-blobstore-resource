#!/bin/bash -eu

ROOT=${PWD}

function main() {
  if [[ ${APPEND_TIMESTAMP_ON_FILENAME} -eq "1" ]]; then
    CONFIGURATION_FILENAME="${CONFIGURATION_FILENAME}-$(date +%s)"
  fi

  echo "writing to ${CONFIGURATION_FILENAME}..."
  echo "The date is $(date)" \
    > "${ROOT}/configuration/${CONFIGURATION_FILENAME}"
}

main
