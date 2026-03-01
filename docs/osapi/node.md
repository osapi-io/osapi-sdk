# NodeService

Node management, network configuration, and command execution. This
is the largest service -- it combines node info, network, and command
operations that all target a specific host.

## Methods

### Node Info

| Method | Description |
| ------ | ----------- |
| `Status(ctx, target)` | Full node status (OS, disk, memory, load) |
| `Hostname(ctx, target)` | Get system hostname |
| `Disk(ctx, target)` | Get disk usage |
| `Memory(ctx, target)` | Get memory statistics |
| `Load(ctx, target)` | Get load averages |
| `OS(ctx, target)` | Get operating system info |
| `Uptime(ctx, target)` | Get uptime |

### Network

| Method | Description |
| ------ | ----------- |
| `GetDNS(ctx, target, iface)` | Get DNS config for an interface |
| `UpdateDNS(ctx, target, iface, servers, search)` | Update DNS servers |
| `Ping(ctx, target, address)` | Ping a host |

### Command

| Method | Description |
| ------ | ----------- |
| `Exec(ctx, req)` | Execute a command directly (no shell) |
| `Shell(ctx, req)` | Execute via `/bin/sh -c` (pipes, redirects) |

## Usage

```go
// Get hostname
resp, err := client.Node.Hostname(ctx, "_any")

// Get disk usage from all hosts
resp, err := client.Node.Disk(ctx, "_all")

// Update DNS
resp, err := client.Node.UpdateDNS(
    ctx, "web-01", "eth0",
    []string{"8.8.8.8", "8.8.4.4"},
    nil,
)

// Execute a command
resp, err := client.Node.Exec(ctx, osapi.ExecRequest{
    Command: "apt",
    Args:    []string{"install", "-y", "nginx"},
    Target:  "_all",
})

// Execute a shell command
resp, err := client.Node.Shell(ctx, osapi.ShellRequest{
    Command: "ps aux | grep nginx",
    Target:  "_any",
})
```

## Permissions

Node info requires `node:read`. Network read requires `network:read`.
DNS updates require `network:write`. Commands require
`command:execute`.
