#!/bin/bash

# Exit on error
set -e

IMAGE_NAME="sindrema/auto-secret-operator"
VERSION_FILE="VERSION"

# Read current version
if [ ! -f "$VERSION_FILE" ]; then
    echo "0.1.0" > "$VERSION_FILE"
fi

CURRENT_VERSION=$(cat "$VERSION_FILE")
echo "Current version: $CURRENT_VERSION"

# Parse version
IFS='.' read -r -a version_parts <<< "$CURRENT_VERSION"
MAJOR="${version_parts[0]}"
MINOR="${version_parts[1]}"
PATCH="${version_parts[2]}"

# Increment patch version
PATCH=$((PATCH + 1))
VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "New version: $VERSION"

# Update VERSION file
echo "$VERSION" > "$VERSION_FILE"

echo "Building multi-architecture Docker image with version: $VERSION"

# Create buildx builder if it doesn't exist
docker buildx create --name multiarch-builder --use 2>/dev/null || docker buildx use multiarch-builder || true

# Build and push multi-architecture image (amd64 and arm64)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ${IMAGE_NAME}:${VERSION} \
    -t ${IMAGE_NAME}:latest \
    --push \
    .

echo "Successfully built and pushed:"
echo "  - ${IMAGE_NAME}:${VERSION}"
echo "  - ${IMAGE_NAME}:latest"
