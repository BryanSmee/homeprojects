#!/usr/bin/env bash
# Computes the GHCR image reference(s) and whether to push, writing them to
# GITHUB_OUTPUT. Publishes on pushes to main (datetime version + latest) and on
# tag/release events (the repo tag); PRs build only.
set -euo pipefail

IMAGE="ghcr.io/${GITHUB_REPOSITORY,,}${IMAGE_SUFFIX:-}"

if [[ "$GITHUB_EVENT_NAME" == "pull_request" ]]; then
  echo "push=false" >> "$GITHUB_OUTPUT"
  echo "tags=$IMAGE:pr" >> "$GITHUB_OUTPUT"
elif [[ "$GITHUB_EVENT_NAME" == "release" ]]; then
  echo "push=true" >> "$GITHUB_OUTPUT"
  echo "tags=$IMAGE:${RELEASE_TAG}" >> "$GITHUB_OUTPUT"
elif [[ "$GITHUB_REF_TYPE" == "tag" ]]; then
  echo "push=true" >> "$GITHUB_OUTPUT"
  echo "tags=$IMAGE:${GITHUB_REF_NAME}" >> "$GITHUB_OUTPUT"
else
  VERSION="$(date -u +'%Y%m%d%H%M%S')"
  echo "push=true" >> "$GITHUB_OUTPUT"
  echo "tags=$IMAGE:${VERSION},$IMAGE:latest" >> "$GITHUB_OUTPUT"
fi
