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

echo "Building Docker image with version: $VERSION"

# Build the Docker image with version tag
docker build -t ${IMAGE_NAME}:${VERSION} .

# Tag the image as latest
docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest

echo "Pushing Docker images to registry..."

# Push version tag
docker push ${IMAGE_NAME}:${VERSION}

# Push latest tag
docker push ${IMAGE_NAME}:latest

echo "Successfully built and pushed:"
echo "  - ${IMAGE_NAME}:${VERSION}"
echo "  - ${IMAGE_NAME}:latest"
