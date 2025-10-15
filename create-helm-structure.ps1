# Create Helm chart directory structure
$helmDir = "C:\Work\db-secret-operator\helm\auto-secret-operator"
$templatesDir = "$helmDir\templates"
$crdsDir = "$templatesDir\crds"

# Create directories
New-Item -ItemType Directory -Force -Path $crdsDir | Out-Null

# Copy CRD files
Copy-Item "C:\Work\db-secret-operator\deploy\crds\*.yaml" -Destination $crdsDir

Write-Host "Helm chart directory structure created successfully!"
Write-Host "Directory: $helmDir"
