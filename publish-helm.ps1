# GitHub Pages Publishing Script
# This script packages the Helm chart and publishes it to GitHub Pages

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

Write-Host "Publishing Helm Chart to GitHub Pages..." -ForegroundColor Green

$baseDir = $PSScriptRoot
$helmDir = Join-Path $baseDir "helm\auto-secret-operator"
$pagesDir = Join-Path $baseDir "gh-pages"

# Ensure Helm chart exists
if (-not (Test-Path $helmDir)) {
    Write-Host "Helm chart not found. Running setup script..." -ForegroundColor Yellow
    & (Join-Path $baseDir "setup-helm-chart.ps1")
}

# Read version from Chart.yaml if not provided
if ([string]::IsNullOrEmpty($Version)) {
    $chartYaml = Get-Content (Join-Path $helmDir "Chart.yaml") -Raw
    if ($chartYaml -match "version:\s*(.+)") {
        $Version = $Matches[1].Trim()
        Write-Host "Using version from Chart.yaml: $Version" -ForegroundColor Cyan
    } else {
        Write-Host "Error: Could not determine version" -ForegroundColor Red
        exit 1
    }
}

# Create gh-pages directory if it doesn't exist
if (-not (Test-Path $pagesDir)) {
    New-Item -ItemType Directory -Path $pagesDir | Out-Null
}

# Package the Helm chart
Write-Host "Packaging Helm chart..." -ForegroundColor Yellow
helm package $helmDir --destination $pagesDir

# Generate or update the index
Write-Host "Updating Helm repository index..." -ForegroundColor Yellow
if (Test-Path (Join-Path $pagesDir "index.yaml")) {
    helm repo index $pagesDir --url https://sindrema.github.io/auto-secret-operator --merge (Join-Path $pagesDir "index.yaml")
} else {
    helm repo index $pagesDir --url https://sindrema.github.io/auto-secret-operator
}

Write-Host "`nHelm chart packaged successfully!" -ForegroundColor Green
Write-Host "Chart: $pagesDir\auto-secret-operator-$Version.tgz" -ForegroundColor Cyan

Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Commit and push the gh-pages directory:" -ForegroundColor White
Write-Host "   git add gh-pages/" -ForegroundColor Gray
Write-Host "   git commit -m 'Publish Helm chart v$Version'" -ForegroundColor Gray
Write-Host "   git push origin main" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Enable GitHub Pages for this repository:" -ForegroundColor White
Write-Host "   - Go to Settings > Pages" -ForegroundColor Gray
Write-Host "   - Set Source to 'Deploy from a branch'" -ForegroundColor Gray
Write-Host "   - Select 'main' branch and '/gh-pages' folder" -ForegroundColor Gray
Write-Host "   - Or use GitHub Actions to deploy (recommended)" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Register on Artifact Hub:" -ForegroundColor White
Write-Host "   - Go to https://artifacthub.io/" -ForegroundColor Gray
Write-Host "   - Sign in with GitHub" -ForegroundColor Gray
Write-Host "   - Add repository: https://sindrema.github.io/auto-secret-operator/index.yaml" -ForegroundColor Gray
