#!/bin/bash -exu

ROOT=${PWD}
GOPATH=${PWD}/go
PATH=${GOPATH}/bin:$PATH

function main() {
  apt-get update
  apt-get install -yyq zip
  go get github.com/onsi/ginkgo/ginkgo

  mkdir -p "${GOPATH}/src/github.com/pivotal-cf"
  ln -s "${ROOT}/azure-blobstore-resource" "${GOPATH}/src/github.com/pivotal-cf/azure-blobstore-resource"
  ginkgo -r "${GOPATH}/src/github.com/pivotal-cf/azure-blobstore-resource"
}

main
