#!/bin/bash -exu

ROOT=${PWD}

function main() {
  pushd "${ROOT}/azure-blobstore-resource" > /dev/null
    ./scripts/build
  popd > /dev/null

  cp -R ${ROOT}/azure-blobstore-resource/* "${ROOT}/workspace"
}

main
