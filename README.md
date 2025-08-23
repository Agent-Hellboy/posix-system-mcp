# POSIX System MCP Server

[![CI](https://github.com/Agent-Hellboy/linux-system-mcp/actions/workflows/makefile.yml/badge.svg)](https://github.com/Agent-Hellboy/linux-system-mcp/actions/workflows/makefile.yml)
[![codecov](https://codecov.io/gh/Agent-Hellboy/linux-system-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/Agent-Hellboy/linux-system-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/Agent-Hellboy/linux-system-mcp)](https://goreportcard.com/report/github.com/Agent-Hellboy/linux-system-mcp)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance Model Context Protocol (MCP) server built with Go that provides POSIX system monitoring capabilities for Claude Desktop, Cursor, and other MCP clients.

## Features

- **System Info**: CPU, memory, disk, network, process monitoring  
- **Real-time**: Live system metrics with minimal latency
- **Lightweight**: Single binary, no external dependencies
- **Cross-platform**: Linux, macOS, Unix variants (POSIX-compliant systems)

## Requirements

- Go 1.21 or higher
- MCP-compatible client (Claude Desktop, Cursor, etc.)

## Installation

```bash
# Clone the repository
git clone https://github.com/Agent-Hellboy/linux-system-mcp.git
cd linux-system-mcp

# Configure system (checks Go installation)
./configure

# Install for Claude Desktop
make install claude

# Install for Cursor  
make install cursor

# Or install for both
make install claude cursor
```

## Available Tools

| Tool | Description |
|------|-------------|
| `get_system_info` | System information (hostname, OS, uptime, etc.) |
| `get_cpu_info` | CPU usage and details |
| `get_memory_info` | Memory and swap usage |
| `get_disk_info` | Disk usage by partition |
| `get_network_info` | Network interface statistics |
| `get_process_info` | Running process information |
| `get_load_average` | System load averages |

## Usage Examples

```
# Get system overview
get_system_info

# Get CPU usage per core
get_cpu_info {"per_cpu": true}

# Get top 10 processes by CPU usage
get_process_info {"limit": 10, "sort_by": "cpu"}

# Get disk usage for root partition
get_disk_info {"path": "/"}
```

## Development

```bash
./configure
make deps      # Install dependencies
make build     # Build binary
make test      # Run tests
make run       # Run server directly
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Pull requests are welcome! Please run `make test` before submitting.