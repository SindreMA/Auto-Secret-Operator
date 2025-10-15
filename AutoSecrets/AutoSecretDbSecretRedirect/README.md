# AutoSecretDbSecretRedirect

The `AutoSecretDbSecretRedirect` custom resource watches an existing Kubernetes Secret containing a database URI and automatically creates a new secret with multiple connection string formats. This is particularly useful when you need to connect to the same database from different application frameworks that require different URI formats.

## Features

- **Automatic URI Transformation**: Converts a PostgreSQL URI into multiple formats
- **Auto-Detection**: Automatically detects when the source secret changes and updates the created secret
- **Multiple Formats**: Generates connection strings for:
  - Microsoft SQL Server style (works with Npgsql .NET provider)
  - ODBC
  - ADO.NET (Npgsql format)
  - JDBC
  - Standard PostgreSQL URI

## Use Case

You have a secret created by `AutoSecretDb` or any other source that contains a `uri` field with a PostgreSQL connection string. Your .NET application needs a Microsoft-style connection string format instead of the standard PostgreSQL URI format.

## How It Works

1. Create an `AutoSecretDbSecretRedirect` resource pointing to your source secret
2. The controller reads the `uri` field from the source secret
3. It parses the URI and transforms it into multiple connection string formats
4. A new secret is created with all the connection string variants
5. When the source secret changes, the redirect secret is automatically updated

## Example

### Source Secret (already exists)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db-readonly
  namespace: mynamespace
data:
  uri: postgresql://myapp-db-user:PASSWORD@postgres-cluster.svc.cluster.local:5432/myapp_db
  # ... other fields
```

### Create the Redirect
```yaml
apiVersion: auto-secret.io/v1alpha1
kind: AutoSecretDbSecretRedirect
metadata:
  name: myapp-db-secret-redirect
  namespace: mynamespace
spec:
  secretname: myapp-db-readonly
  # targetSecretName: my-custom-name  # Optional: defaults to <secretname>-redirect
```

### Result Secret (automatically created)
The controller will create a secret named `myapp-db-readonly-redirect` containing:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db-readonly-redirect
  namespace: mynamespace
data:
  # Original URI
  uri: postgresql://myapp-db-user:PASSWORD@postgres-cluster.svc.cluster.local:5432/myapp_db
  original-uri: postgresql://myapp-db-user:PASSWORD@postgres-cluster.svc.cluster.local:5432/myapp_db

  # Microsoft SQL Server style (Npgsql .NET)
  ms-uri: Server=postgres-cluster.svc.cluster.local;Port=5432;Database=myapp_db;User Id=myapp-db-user;Password=PASSWORD;

  # ODBC
  odbc-uri: Driver={PostgreSQL Unicode};Server=postgres-cluster.svc.cluster.local;Port=5432;Database=myapp_db;Uid=myapp-db-user;Pwd=PASSWORD;

  # ADO.NET (Npgsql)
  adonet-uri: Host=postgres-cluster.svc.cluster.local;Port=5432;Database=myapp_db;Username=myapp-db-user;Password=PASSWORD;

  # JDBC
  jdbc-uri: jdbc:postgresql://postgres-cluster.svc.cluster.local:5432/myapp_db?user=myapp-db-user&password=PASSWORD

  # Individual components
  username: myapp-db-user
  password: PASSWORD
  host: postgres-cluster.svc.cluster.local
  port: "5432"
  dbname: myapp_db
```

## Using in Your Application

### .NET Application (C#)
```csharp
// Read the ms-uri field from the secret
var connectionString = Environment.GetEnvironmentVariable("DB_MS_URI");
using var connection = new NpgsqlConnection(connectionString);
```

### Java Application
```java
// Read the jdbc-uri field from the secret
String jdbcUrl = System.getenv("DB_JDBC_URI");
Connection conn = DriverManager.getConnection(jdbcUrl);
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: DB_MS_URI
          valueFrom:
            secretKeyRef:
              name: myapp-db-readonly-redirect
              key: ms-uri
        - name: DB_JDBC_URI
          valueFrom:
            secretKeyRef:
              name: myapp-db-readonly-redirect
              key: jdbc-uri
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `secretname` | string | Yes | Name of the source secret to watch (must contain a `uri` field) |
| `targetSecretName` | string | No | Name for the created secret (defaults to `<secretname>-redirect`) |

## Status Fields

| Field | Description |
|-------|-------------|
| `targetSecretName` | Name of the created secret |
| `sourceSecretResourceVersion` | Resource version of the source secret that was last processed |
| `conditions` | Standard Kubernetes conditions |

## Auto-Update Behavior

The controller watches the source secret. When the source secret is updated (e.g., password rotation), the redirect secret is automatically updated with the new connection strings. This is tracked using the `sourceSecretResourceVersion` in the status.

## Installation

The CRD is automatically installed when you deploy the db-secret-operator:

```bash
kubectl apply -f deploy/crds/auto-secret.io_autosecretdbsecretredirects.yaml
kubectl apply -f deploy/deployment.yaml
```

## Shortname

You can use the shortname `asdbsr` for convenience:

```bash
kubectl get asdbsr
kubectl describe asdbsr myapp-db-secret-redirect
```
