#!/bin/bash -exu

ROOT=${PWD}

function main() {
  ginkgo -r "${ROOT}/azure-blobstore-resource"
}

main
