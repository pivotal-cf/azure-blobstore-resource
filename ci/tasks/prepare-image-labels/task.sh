#!/bin/bash -exu

ROOT="${PWD}"

function main() {
  local version="$(cat ${ROOT}/version/version)"
  local commit_sha="$(git --git-dir="${ROOT}/azure-blobstore-resource/.git" rev-parse HEAD)"

  cat > "${ROOT}/image-labels/labels.json" <<EOF
{
  "url": "https://github.com/pivotal-cf/azure-blobstore-resource/releases/tag/v${version}",
  "version": "v${version}",
  "commit": "${commit_sha}"
}
EOF
}

main
