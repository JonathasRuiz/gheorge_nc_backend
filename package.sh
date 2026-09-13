#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME=docker.io/jhonnyvennom/gheorge-nc-backend
VERSION=""

usage() {
  echo "Usage: $0 -v <version>"
  echo ""
  echo "Builds and pushes ${IMAGE_NAME} to Docker Hub."
  echo ""
  echo "Options:"
  echo "  -v <version>   Image version tag (e.g. 1.0.0 or v1.0.0)"
  echo "  -h             Show this help"
}

while getopts "v:h" opt; do
  case "$opt" in
    v) VERSION="$OPTARG" ;;
    h) usage; exit 0 ;;
    *) usage >&2; exit 1 ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  echo "Error: version is required (-v)" >&2
  usage >&2
  exit 1
fi

VERSION="${VERSION#v}"
IMAGE="${IMAGE_NAME}:${VERSION}"
IMAGE_LATEST="${IMAGE_NAME}:latest"

echo "==> Building ${IMAGE}"
podman build -t "$IMAGE" -t "$IMAGE_LATEST" .

echo "==> Pushing ${IMAGE}"
podman push "$IMAGE"
podman push "$IMAGE_LATEST"

echo "==> Done: ${IMAGE}"
