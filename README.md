# Linux System MCP Server

[![CI](https://github.com/Agent-Hellboy/linux-system-mcp/actions/workflows/makefile.yml/badge.svg)](https://github.com/Agent-Hellboy/linux-system-mcp/actions/workflows/makefile.yml)
[![codecov](https://codecov.io/gh/Agent-Hellboy/linux-system-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/Agent-Hellboy/linux-system-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/Agent-Hellboy/linux-system-mcp)](https://goreportcard.com/report/github.com/Agent-Hellboy/linux-system-mcp)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)


A high-performance Model Context Protocol (MCP) server built with Go that provides comprehensive Linux system monitoring capabilities. This server allows Claude Desktop, Cursor, and other MCP-compatible clients to access real-time system information including CPU usage, memory statistics, disk usage, network statistics, process information, and more.

## 🚀 Features

### System Information
- **Host Details**: Hostname, OS, platform, kernel version, architecture
- **Uptime & Boot Time**: System uptime and boot timestamp
- **Hardware Info**: CPU model, core counts, cache size, flags
- **Virtualization**: Detection of virtualization systems and roles
- **Temperature Sensors**: Hardware temperature monitoring (when available)

### Performance Monitoring
- **CPU Usage**: Real-time CPU utilization per core or aggregated
- **Memory Stats**: RAM and swap usage with detailed breakdown
- **Load Average**: 1, 5, and 15-minute load averages
- **Process Monitoring**: Detailed process information with sorting and filtering

### Storage & Network
- **Disk Usage**: File system usage for all mounted volumes
- **Network Statistics**: Interface statistics including bytes, packets, errors
- **I/O Counters**: Network interface input/output counters

### Advanced Process Management
- **Process Details**: PID, CPU/memory usage, status, threads, command line
- **Process Filtering**: Filter by name, PID, or other criteria
- **Process Sorting**: Sort by CPU usage, memory, PID, or name
- **User Information**: Process owner details

## 📋 Requirements

- **Go**: 1.21 or higher
- **Operating System**: Linux, macOS, Windows (optimized for Linux)
- **MCP Client**: Claude Desktop, Cursor, or any MCP-compatible application

## 🔧 Installation

### Quick Install

```bash
# Clone the repository
git clone https://github.com/Agent-Hellboy/linux-system-mcp.git
cd linux-system-mcp

# Install for Claude Desktop
make install claude

# Install for Cursor
make install cursor

```

### Manual Build

```bash
# Build the binary
make build

# Create release package
make release
```

### Development Setup

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run server directly
make run
```

## 🎯 Usage

### Available Tools

The MCP server provides the following tools:

| Tool | Description | Parameters |
|------|-------------|-----------|
| `get_system_info` | Get comprehensive system information | None |
| `get_cpu_info` | Get detailed CPU information and usage | `per_cpu`: boolean (optional) |
| `get_memory_info` | Get memory and swap usage information | None |
| `get_disk_info` | Get disk usage for mounted filesystems | `path`: string (optional) |
| `get_network_info` | Get network interface statistics | `interface`: string (optional) |
| `get_process_info` | Get running process information | `pid`, `name`, `limit`, `sort_by` (all optional) |
| `get_load_average` | Get system load averages | None |

### Example Usage in Claude Desktop

```
# Get system overview
get_system_info

# Get CPU usage per core
get_cpu_info {"per_cpu": true}

# Get disk usage for root partition
get_disk_info {"path": "/"}

# Get top 10 processes by CPU usage
get_process_info {"limit": 10, "sort_by": "cpu"}

# Get processes matching a name
get_process_info {"name": "python", "limit": 5}

# Get network stats for specific interface
get_network_info {"interface": "eth0"}
```

## ⚙️ Configuration

### Claude Desktop Configuration

The installation automatically configures Claude Desktop by updating:
`~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "linux-system-monitor": {
      "command": "/Users/username/.local/bin/linux-system-mcp",
      "args": [],
      "env": {},
      "description": "High-performance Linux system monitoring MCP server built with Go"
    }
  }
}
```

### Cursor Configuration

For Cursor, the configuration is added to:
`~/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "linux-system-monitor": {
      "command": "/Users/username/.local/bin/linux-system-mcp",
      "args": [],
      "env": {},
      "description": "High-performance Linux system monitoring MCP server built with Go"
    }
  }
}
```

## 🛠️ Development

### Building

```bash
# Development build
go build -o bin/linux-system-mcp .

# Production build with version
go build -o bin/linux-system-mcp -ldflags="-X main.Version=1.0.0" .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o bin/linux-system-mcp-linux .
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...
```

## 📊 Monitoring Capabilities

### System Metrics

- **CPU**: Usage percentages, core counts, model information, cache sizes
- **Memory**: Total, available, used, cached, buffered, swap statistics  
- **Disk**: Usage by mount point, filesystem types, inode usage
- **Network**: Interface statistics, bytes/packets sent/received, error rates
- **Processes**: CPU/memory usage, thread counts, user ownership, command lines
- **Load**: System load averages over different time periods

### Performance Features

- **Efficient**: Built with Go for high performance and low memory usage
- **Real-time**: Live system metrics with minimal latency
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Lightweight**: Single binary with no external dependencies
- **Concurrent**: Handles multiple requests efficiently

## 🔍 Troubleshooting

### Common Issues

**Server not starting:**
```bash
# Check if binary is installed and executable
ls -la ~/.local/bin/linux-system-mcp

# Test server directly
~/.local/bin/linux-system-mcp --version
```

**Configuration issues:**
```bash
# Check installation status
make status

# Reinstall configuration
make install claude cursor
```

**Permission errors:**
```bash
# Ensure binary is executable
chmod +x ~/.local/bin/linux-system-mcp

# Check PATH includes ~/.local/bin
echo $PATH | grep -o ~/.local/bin
```

### Debugging

```bash
# Run server with debug output
DEBUG=1 ~/.local/bin/linux-system-mcp

# Test individual tools
echo '{"method": "tools/list"}' | ~/.local/bin/linux-system-mcp
```

## 📈 Performance

- **Memory Usage**: ~10MB typical memory footprint
- **CPU Overhead**: <1% CPU usage during monitoring
- **Response Time**: <100ms for most queries
- **Concurrent Users**: Supports multiple simultaneous connections
- **Data Accuracy**: Real-time system metrics with high precision

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests (`make test`)
6. Format code (`make fmt`)
7. Commit changes (`git commit -m 'Add amazing feature'`)
8. Push to branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gopsutil](https://github.com/shirou/gopsutil) - Cross-platform system monitoring library
- [Model Context Protocol](https://modelcontextprotocol.io) - Protocol specification
- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) - Go SDK for MCP servers

## 📞 Support

If you have questions or need help:

- 📝 [Open an Issue](https://github.com/Agent-Hellboy/linux-system-mcp/issues)
- 💬 [GitHub Discussions](https://github.com/Agent-Hellboy/linux-system-mcp/discussions)
- 📧 Contact: [princekrroshan01@gmail.com]

---

**Built with ❤️ using Go and the Model Context Protocol**
