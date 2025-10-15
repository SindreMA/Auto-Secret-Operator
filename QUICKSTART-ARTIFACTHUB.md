# Quick Start Guide - Publishing to Artifact Hub

## TL;DR

1. **Run setup once:**
   ```bash
   ./setup-helm-chart.ps1
   ```

2. **Create and push a version tag:**
   ```bash
   git tag v0.1.12
   git push origin v0.1.12
   ```

3. **GitHub Actions will automatically:**
   - Package your Helm chart
   - Publish to GitHub Pages
   - Create a GitHub Release

4. **Register on Artifact Hub (one-time):**
   - Go to https://artifacthub.io/
   - Sign in with GitHub
   - Add repository: `https://sindrema.github.io/auto-secret-operator/index.yaml`

5. **Done!** Your chart is now discoverable on Artifact Hub.

---

## What Got Created

### Metadata Files
- **artifacthub-pkg.yml** - Package metadata (version, description, CRDs, etc.)
- **artifacthub-repo.yml** - Repository owner information

### Scripts
- **setup-helm-chart.ps1** - Creates complete Helm chart structure
- **publish-helm.ps1** - Manual publishing script (if not using GitHub Actions)

### GitHub Actions
- **.github/workflows/release-helm.yml** - Automated release workflow

### Documentation
- **ARTIFACTHUB-PUBLISHING.md** - Detailed publishing instructions
- **HELM-INSTALL.md** - User installation guide

---

## For Users Installing Your Chart

After publishing, users can install with:

```bash
helm repo add auto-secret-operator https://sindrema.github.io/auto-secret-operator
helm repo update
helm install auto-secret-operator auto-secret-operator/auto-secret-operator
```

Or search on Artifact Hub:
- https://artifacthub.io/

---

## Updating Your Chart

When making changes:

1. Update version in `helm/auto-secret-operator/Chart.yaml`
2. Update `artifacthub-pkg.yml` with changelog
3. Commit changes
4. Create new tag: `git tag v0.1.13`
5. Push: `git push origin v0.1.13`
6. GitHub Actions handles the rest!

---

## Important Notes

- **First time setup**: Enable GitHub Pages in repository Settings > Pages > Source: GitHub Actions
- **Email**: Update your email in `artifacthub-repo.yml` and `artifacthub-pkg.yml`
- **Logo**: Optionally add a logo URL to `artifacthub-pkg.yml`
- **Wait time**: Artifact Hub indexes every 30 minutes

---

See **ARTIFACTHUB-PUBLISHING.md** for complete details.
