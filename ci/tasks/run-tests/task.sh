#!/bin/bash -exu

ROOT=${PWD}

function main() {
  apt-get update
  apt-get install -yyq zip
  go get github.com/onsi/ginkgo/ginkgo

  pushd "${ROOT}/azure-blobstore-resource" > /dev/null
    ginkgo -r .
  popd > /dev/null
}

main
