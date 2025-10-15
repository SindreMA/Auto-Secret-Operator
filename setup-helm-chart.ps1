# Setup Helm Chart
# Run this script to create the Helm chart structure

Write-Host "Creating Helm chart structure..." -ForegroundColor Green

# Define directories
$baseDir = $PSScriptRoot
$helmDir = Join-Path $baseDir "helm\auto-secret-operator"
$templatesDir = Join-Path $helmDir "templates"
$crdsDir = Join-Path $templatesDir "crds"
$workflowsDir = Join-Path $baseDir ".github\workflows"

# Create directory structure
New-Item -ItemType Directory -Force -Path $crdsDir | Out-Null
New-Item -ItemType Directory -Force -Path $workflowsDir | Out-Null

# Copy CRD files
$sourcecrds = Join-Path $baseDir "deploy\crds"
Copy-Item "$sourceCrds\*.yaml" -Destination $crdsDir -Force

# Create Chart.yaml
$chartYaml = @'
apiVersion: v2
name: auto-secret-operator
description: A Kubernetes operator that automatically generates secrets with passwords, database credentials, and GUIDs
type: application
version: 0.1.0
appVersion: "0.1.15"
keywords:
  - kubernetes
  - operator
  - secrets
  - passwords
  - database
maintainers:
  - name: sindrema
home: https://github.com/sindrema/auto-secret-operator
'@
$chartYaml | Out-File -FilePath (Join-Path $helmDir "Chart.yaml") -Encoding utf8 -NoNewline

# Create values.yaml
$valuesYaml = @'
# Default values for auto-secret-operator

replicaCount: 1

image:
  repository: sindrema/auto-secret-operator
  pullPolicy: IfNotPresent
  # Overrides the image tag whose default is the chart appVersion
  tag: ""

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

# Namespace to deploy the operator
namespace: auto-secret-operator

serviceAccount:
  # Specifies whether a service account should be created
  create: true
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name: ""

podAnnotations: {}

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true

# Operator configuration
operator:
  # Enable leader election for controller manager
  leaderElect: true
  # Metrics and health probe ports
  metricsPort: 8080
  healthPort: 8081

resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi

livenessProbe:
  httpGet:
    path: /healthz
    port: health
  initialDelaySeconds: 15
  periodSeconds: 20

readinessProbe:
  httpGet:
    path: /readyz
    port: health
  initialDelaySeconds: 5
  periodSeconds: 10

nodeSelector: {}

tolerations: []

affinity: {}

# RBAC configuration
rbac:
  # Specifies whether RBAC resources should be created
  create: true
'@
$valuesYaml | Out-File -FilePath (Join-Path $helmDir "values.yaml") -Encoding utf8 -NoNewline

# Create _helpers.tpl
$helpersTpl = @'
{{/*
Expand the name of the chart.
*/}}
{{- define "auto-secret-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "auto-secret-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "auto-secret-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "auto-secret-operator.labels" -}}
helm.sh/chart: {{ include "auto-secret-operator.chart" . }}
{{ include "auto-secret-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "auto-secret-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "auto-secret-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "auto-secret-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "auto-secret-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image tag
*/}}
{{- define "auto-secret-operator.imageTag" -}}
{{- default .Chart.AppVersion .Values.image.tag }}
{{- end }}
'@
$helpersTpl | Out-File -FilePath (Join-Path $templatesDir "_helpers.tpl") -Encoding utf8 -NoNewline

# Create namespace.yaml
$namespaceYaml = @'
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.namespace }}
  labels:
    {{- include "auto-secret-operator.labels" . | nindent 4 }}
'@
$namespaceYaml | Out-File -FilePath (Join-Path $templatesDir "namespace.yaml") -Encoding utf8 -NoNewline

# Create serviceaccount.yaml
$serviceaccountYaml = @'
{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "auto-secret-operator.serviceAccountName" . }}
  namespace: {{ .Values.namespace }}
  labels:
    {{- include "auto-secret-operator.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
'@
$serviceaccountYaml | Out-File -FilePath (Join-Path $templatesDir "serviceaccount.yaml") -Encoding utf8 -NoNewline

# Create clusterrole.yaml
$clusterroleYaml = @'
{{- if .Values.rbac.create -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ include "auto-secret-operator.fullname" . }}-role
  labels:
    {{- include "auto-secret-operator.labels" . | nindent 4 }}
rules:
# AutoSecretBasic permissions
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretbasics
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretbasics/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretbasics/finalizers
  verbs:
  - update
# AutoSecretDb permissions
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbs
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbs/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbs/finalizers
  verbs:
  - update
# AutoSecretGuid permissions
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretguids
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretguids/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretguids/finalizers
  verbs:
  - update
# AutoSecretDbSecretRedirect permissions
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbsecretredirects
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbsecretredirects/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - auto-secret.io
  resources:
  - autosecretdbsecretredirects/finalizers
  verbs:
  - update
# Secret permissions
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
# Leader election permissions
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
# Event permissions
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
{{- end }}
'@
$clusterroleYaml | Out-File -FilePath (Join-Path $templatesDir "clusterrole.yaml") -Encoding utf8 -NoNewline

# Create clusterrolebinding.yaml
$clusterrolebindingYaml = @'
{{- if .Values.rbac.create -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ include "auto-secret-operator.fullname" . }}-rolebinding
  labels:
    {{- include "auto-secret-operator.labels" . | nindent 4 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ include "auto-secret-operator.fullname" . }}-role
subjects:
- kind: ServiceAccount
  name: {{ include "auto-secret-operator.serviceAccountName" . }}
  namespace: {{ .Values.namespace }}
{{- end }}
'@
$clusterrolebindingYaml | Out-File -FilePath (Join-Path $templatesDir "clusterrolebinding.yaml") -Encoding utf8 -NoNewline

# Create deployment.yaml
$deploymentYaml = @'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "auto-secret-operator.fullname" . }}
  namespace: {{ .Values.namespace }}
  labels:
    {{- include "auto-secret-operator.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "auto-secret-operator.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "auto-secret-operator.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "auto-secret-operator.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
      - name: manager
        securityContext:
          {{- toYaml .Values.securityContext | nindent 10 }}
        image: "{{ .Values.image.repository }}:{{ include "auto-secret-operator.imageTag" . }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        command:
        - /manager
        args:
        {{- if .Values.operator.leaderElect }}
        - --leader-elect
        {{- end }}
        ports:
        - containerPort: {{ .Values.operator.metricsPort }}
          name: metrics
          protocol: TCP
        - containerPort: {{ .Values.operator.healthPort }}
          name: health
          protocol: TCP
        livenessProbe:
          {{- toYaml .Values.livenessProbe | nindent 10 }}
        readinessProbe:
          {{- toYaml .Values.readinessProbe | nindent 10 }}
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
'@
$deploymentYaml | Out-File -FilePath (Join-Path $templatesDir "deployment.yaml") -Encoding utf8 -NoNewline

# Create NOTES.txt
$notesTxt = @'
Thank you for installing {{ .Chart.Name }}!

Your release is named {{ .Release.Name }}.

The Auto Secret Operator has been deployed to namespace: {{ .Values.namespace }}

To verify the deployment:
  kubectl get pods -n {{ .Values.namespace }}

To create an AutoSecretBasic:
  kubectl apply -f - <<EOF
  apiVersion: auto-secret.io/v1alpha1
  kind: AutoSecretBasic
  metadata:
    name: my-basic-secret
    namespace: default
  spec:
    username: myuser
  EOF

To create an AutoSecretDb:
  kubectl apply -f - <<EOF
  apiVersion: auto-secret.io/v1alpha1
  kind: AutoSecretDb
  metadata:
    name: my-db-secret
    namespace: default
  spec:
    username: dbuser
    dbname: mydb
    dbhost: postgres.default.svc.cluster.local
  EOF

To create an AutoSecretGuid:
  kubectl apply -f - <<EOF
  apiVersion: auto-secret.io/v1alpha1
  kind: AutoSecretGuid
  metadata:
    name: my-guid-secret
    namespace: default
  spec:
    format: uuidv4
  EOF

For more information, visit: https://github.com/sindrema/auto-secret-operator
'@
$notesTxt | Out-File -FilePath (Join-Path $templatesDir "NOTES.txt") -Encoding utf8 -NoNewline

# Create GitHub Actions workflow for releasing
$releaseWorkflow = @'
name: Release Helm Chart

on:
  push:
    tags:
      - 'v*.*.*'
  workflow_dispatch:

permissions:
  contents: write
  pages: write
  id-token: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config user.name "$GITHUB_ACTOR"
          git config user.email "$GITHUB_ACTOR@users.noreply.github.com"

      - name: Install Helm
        uses: azure/setup-helm@v4
        with:
          version: 'latest'

      - name: Setup Helm Chart
        run: |
          chmod +x setup-helm-chart.ps1
          pwsh ./setup-helm-chart.ps1

      - name: Run chart-releaser
        uses: helm/chart-releaser-action@v1.6.0
        env:
          CR_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
        with:
          charts_dir: helm
          skip_existing: true

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        if: startsWith(github.ref, 'refs/tags/')
        with:
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
'@
$releaseWorkflow | Out-File -FilePath (Join-Path $workflowsDir "release-helm.yml") -Encoding utf8 -NoNewline

Write-Host "`nHelm chart created successfully!" -ForegroundColor Green
Write-Host "Location: $helmDir" -ForegroundColor Cyan
Write-Host "GitHub Actions: $workflowsDir" -ForegroundColor Cyan
Write-Host "`nTo install the chart:" -ForegroundColor Yellow
Write-Host "  helm install auto-secret-operator ./helm/auto-secret-operator" -ForegroundColor White
Write-Host "`nTo package the chart:" -ForegroundColor Yellow
Write-Host "  helm package ./helm/auto-secret-operator" -ForegroundColor White
Write-Host "`nTo publish to Artifact Hub:" -ForegroundColor Yellow
Write-Host "  See ARTIFACTHUB-PUBLISHING.md for detailed instructions" -ForegroundColor White
