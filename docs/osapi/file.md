# FileService

File management operations for the OSAPI Object Store. Upload, list, inspect,
and delete files that can be deployed to agents via `Node.FileDeploy`.

## Methods

### Object Store

| Method              | Description                         |
| ------------------- | ----------------------------------- |
| `Upload(ctx, n, d)` | Upload file content to Object Store |
| `List(ctx)`         | List all stored files               |
| `Get(ctx, name)`    | Get file metadata by name           |
| `Delete(ctx, name)` | Delete a file from Object Store     |

### Node File Operations

File deploy and status methods live on `NodeService` because they target a
specific host:

| Method                          | Description                         |
| ------------------------------- | ----------------------------------- |
| `FileDeploy(ctx, opts)`         | Deploy file to agent with SHA check |
| `FileStatus(ctx, target, path)` | Check deployed file status          |

## FileDeployOpts

| Field         | Type           | Required | Description                              |
| ------------- | -------------- | -------- | ---------------------------------------- |
| `ObjectName`  | string         | Yes      | Name of the file in the Object Store     |
| `Path`        | string         | Yes      | Destination path on the target host      |
| `ContentType` | string         | Yes      | `"raw"` or `"template"`                  |
| `Mode`        | string         | No       | File permission mode (e.g., `"0644"`)    |
| `Owner`       | string         | No       | File owner user                          |
| `Group`       | string         | No       | File owner group                         |
| `Vars`        | map[string]any | No       | Template variables for `"template"` type |
| `Target`      | string         | Yes      | Host target (see Targeting below)        |

## Usage

```go
// Upload a file
resp, err := client.File.Upload(ctx, "nginx.conf", configBytes)

// List all files
resp, err := client.File.List(ctx)

// Get file metadata
resp, err := client.File.Get(ctx, "nginx.conf")

// Delete a file
resp, err := client.File.Delete(ctx, "nginx.conf")

// Deploy a raw file to a specific host
resp, err := client.Node.FileDeploy(ctx, osapi.FileDeployOpts{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Owner:       "root",
    Group:       "root",
    Target:      "web-01",
})

// Deploy a template file with variables
resp, err := client.Node.FileDeploy(ctx, osapi.FileDeployOpts{
    ObjectName:  "app.conf.tmpl",
    Path:        "/etc/app/config.yaml",
    ContentType: "template",
    Vars: map[string]any{
        "port":  8080,
        "debug": false,
    },
    Target: "_all",
})

// Check file status on a host
resp, err := client.Node.FileStatus(
    ctx, "web-01", "/etc/nginx/nginx.conf",
)
```

## Targeting

`FileDeploy` and `FileStatus` accept any valid target: `_any`, `_all`, a
hostname, or a label selector (`key:value`).

Object Store operations (`Upload`, `List`, `Get`, `Delete`) are server-side and
do not use targeting.

## Idempotency

`FileDeploy` compares the SHA-256 of the Object Store content against the
deployed file. If the hashes match, the file is not rewritten and the response
reports `Changed: false`.

## Permissions

Object Store operations require `file:read` (list, get) or `file:write` (upload,
delete). Deploy requires `file:write`. Status requires `file:read`.
