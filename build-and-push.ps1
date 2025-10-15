# Exit on error
$ErrorActionPreference = "Stop"

$IMAGE_NAME = "sindrema/auto-secret-operator"
$VERSION_FILE = "VERSION"

# Read current version
if (-not (Test-Path $VERSION_FILE)) {
    "0.1.0" | Out-File -FilePath $VERSION_FILE -NoNewline -Encoding utf8
}

$CURRENT_VERSION = (Get-Content $VERSION_FILE -Raw).Trim()
Write-Host "Current version: $CURRENT_VERSION"

# Parse version
$version_parts = $CURRENT_VERSION.Split('.')
$MAJOR = $version_parts[0]
$MINOR = $version_parts[1]
$PATCH = [int]$version_parts[2]

# Increment patch version
$PATCH = $PATCH + 1
$VERSION = "${MAJOR}.${MINOR}.${PATCH}"

Write-Host "New version: $VERSION"

# Update VERSION file
$VERSION | Out-File -FilePath $VERSION_FILE -NoNewline -Encoding utf8

Write-Host "Building multi-architecture Docker image with version: $VERSION"

# Create buildx builder if it doesn't exist
try {
    docker buildx create --name multiarch-builder --use 2>$null
} catch {
    try {
        docker buildx use multiarch-builder
    } catch {
        # Builder already exists and is in use
    }
}

# Build and push multi-architecture image (amd64 and arm64)
docker buildx build `
    --platform linux/amd64,linux/arm64 `
    -t ${IMAGE_NAME}:${VERSION} `
    -t ${IMAGE_NAME}:latest `
    --push `
    .

Write-Host "Successfully built and pushed:"
Write-Host "  - ${IMAGE_NAME}:${VERSION}"
Write-Host "  - ${IMAGE_NAME}:latest"
