#!/bin/bash -exu

ROOT=${PWD}

function main() {
  pushd "${ROOT}/azure-blobstore-resource" > /dev/null
    ./scripts/build
    echo "release_candidate" > tag
  popd > /dev/null
}

main
