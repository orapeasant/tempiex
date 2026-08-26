#!/usr/bin/env bash
# scripts/docker-build.sh — build all Docker images from repo root
# Usage: ./scripts/docker-build.sh [tag]
set -euo pipefail

TAG="${1:-latest}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Building from context: $REPO_ROOT  tag=:$TAG"

docker build \
  -f tempiex/Dockerfile \
  -t tempiex/core:$TAG \
  "$REPO_ROOT"

docker build \
  -f server/Dockerfile \
  -t tempiex/server:$TAG \
  "$REPO_ROOT"

docker build \
  -f cli/Dockerfile \
  -t tempiex/cli:$TAG \
  "$REPO_ROOT"

docker build \
  -f sdk-python/Dockerfile \
  -t tempiex/sdk-python:$TAG \
  "$REPO_ROOT"

echo "All images built successfully."
