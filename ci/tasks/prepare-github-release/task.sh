#!/bin/bash -exu

ROOT="${PWD}"

function main() {
  generate_name
  generate_body
}

function generate_name() {
  echo "v$(cat "${ROOT}/version/version")" > "${ROOT}/github-release/name"
}

function generate_body() {
  cat > "${ROOT}/github-release/body" <<EOF
#### Image
[pcfabr/azure-blobstore-resource](https://hub.docker.com/r/pcfabr/azure-blobstore-resource)
_digest: $(cat "${ROOT}/azure-blobstore-resource-final-image/digest")_

#### Base Image
[cloudfoundry/run:tiny](https://hub.docker.com/r/cloudfoundry/run/tags?name=tiny)
_digest: $(cat "${ROOT}/base-image/digest")_
EOF
}

main
