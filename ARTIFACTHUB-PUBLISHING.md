# Publishing to Artifact Hub

This guide explains how to publish the Auto Secret Operator Helm chart to [Artifact Hub](https://artifacthub.io/).

## Prerequisites

- GitHub account with write access to the repository
- Artifact Hub account (sign up at https://artifacthub.io/)

## Method 1: Automatic Publishing via GitHub Actions (Recommended)

### Step 1: Enable GitHub Pages

1. Go to your repository on GitHub
2. Navigate to **Settings** > **Pages**
3. Under **Build and deployment**:
   - Source: Select **GitHub Actions**

### Step 2: Create a Release

The GitHub Actions workflow (`.github/workflows/release-helm.yml`) will automatically:
- Package the Helm chart
- Create a GitHub release
- Publish to GitHub Pages

To trigger a release:

```bash
# Tag the version
git tag v0.1.15
git push origin v0.1.15
```

Or trigger manually:
- Go to **Actions** tab in GitHub
- Select **Release Helm Chart** workflow
- Click **Run workflow**

### Step 3: Register on Artifact Hub

1. Go to https://artifacthub.io/
2. Click **Sign in** and authenticate with GitHub
3. Click on your profile > **Control Panel**
4. Click **Add repository**
5. Fill in the details:
   - **Name**: auto-secret-operator
   - **Display name**: Auto Secret Operator
   - **URL**: `https://sindrema.github.io/auto-secret-operator/index.yaml`
   - **Kind**: Helm charts
6. Click **Add**

Artifact Hub will automatically:
- Index your Helm chart
- Display it in search results
- Track new versions
- Show metadata from `artifacthub-pkg.yml`

## Method 2: Manual Publishing

### Step 1: Setup and Package

```bash
# Create Helm chart structure
./setup-helm-chart.ps1

# Package the chart
./publish-helm.ps1
```

### Step 2: Commit and Push

```bash
git add gh-pages/
git commit -m "Publish Helm chart v0.1.15"
git push origin main
```

### Step 3: Enable GitHub Pages

1. Go to **Settings** > **Pages**
2. Set Source to **Deploy from a branch**
3. Select **main** branch and **/gh-pages** folder
4. Click **Save**

### Step 4: Register on Artifact Hub

Follow Step 3 from Method 1 above.

## Updating the Chart

When you make changes to the Helm chart:

1. Update the version in `helm/auto-secret-operator/Chart.yaml`
2. Update `artifacthub-pkg.yml` with changes
3. For automatic: Create a new git tag
4. For manual: Run `./publish-helm.ps1` and commit

## Metadata Files

### artifacthub-pkg.yml

This file provides metadata for Artifact Hub:
- Package information
- CRDs
- Installation instructions
- Maintainer details
- Links and documentation

Update this file when you make significant changes.

### artifacthub-repo.yml

This file identifies the repository owner and is used by Artifact Hub to associate the repository with your account.

## Verification

After publishing, verify your chart is available:

```bash
# Add your Helm repository
helm repo add auto-secret-operator https://sindrema.github.io/auto-secret-operator

# Update repositories
helm repo update

# Search for your chart
helm search repo auto-secret-operator

# Install from repository
helm install my-operator auto-secret-operator/auto-secret-operator
```

## Troubleshooting

### Chart not appearing on Artifact Hub

- Check that `artifacthub-pkg.yml` is in the root of your repository or in the chart directory
- Verify the Helm repository URL is accessible
- Wait up to 30 minutes for Artifact Hub to index new charts
- Check Artifact Hub repository settings for any errors

### GitHub Pages not serving content

- Ensure GitHub Pages is enabled in repository settings
- Check that `gh-pages` folder contains `index.yaml` and `.tgz` files
- Verify the GitHub Actions workflow completed successfully

### Invalid Helm chart

- Run `helm lint ./helm/auto-secret-operator` to check for errors
- Validate `Chart.yaml` syntax
- Ensure all required template files exist

## Resources

- [Artifact Hub Documentation](https://artifacthub.io/docs/)
- [Helm Chart Releaser Action](https://github.com/helm/chart-releaser-action)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
