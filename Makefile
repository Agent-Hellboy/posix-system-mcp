# Linux System MCP Server Makefile

# Show help information
help:
	@echo "Linux System MCP Server Build Tool"
	@echo "===================================="
	@echo ""
	@echo "Usage:"
	@echo "  make <target> [parameters]"
	@echo ""
	@echo "Build targets:"
	@echo "  build              Build the binary"
	@echo "  build-linux        Build for Linux (cross-compile)"
	@echo "  clean              Remove build artifacts"
	@echo "  release            Create release package"
	@echo ""
	@echo "Installation targets:"
	@echo "  install [claude] [cursor]  Install binary and configure specified clients"
	@echo "    Examples:"
	@echo "      make install claude        - Install and configure for Claude Desktop only"
	@echo "      make install cursor        - Install and configure for Cursor only"
	@echo "      make install claude cursor - Install and configure for both clients"
	@echo "  uninstall          Remove binary and configurations"
	@echo ""
	@echo "Development targets:"
	@echo "  run                Run the server directly"
	@echo "  test               Run tests"
	@echo "  deps               Install/update dependencies"
	@echo "  fmt                Format code"
	@echo "  lint               Run linter"
	@echo ""
	@echo "Utility targets:"
	@echo "  status             Show installation status"
	@echo "  help               Show this help message"

BINARY_NAME=linux-system-mcp
VERSION=1.0.0
INSTALL_DIR=$(HOME)/.local/bin
CLAUDE_CONFIG_DIR=$(HOME)/Library/Application Support/Claude
CLAUDE_CONFIG_FILE=$(CLAUDE_CONFIG_DIR)/claude_desktop_config.json
CURSOR_CONFIG_DIR=$(HOME)/.cursor
CURSOR_CONFIG_FILE=$(CURSOR_CONFIG_DIR)/mcp.json

# Build the binary
build:
	go build -o bin/$(BINARY_NAME) -ldflags="-X main.Version=$(VERSION)" .

# Build for Linux (cross-compile from macOS)
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux -ldflags="-X main.Version=$(VERSION)" .

# Install binary and configure specified clients (usage: make install claude cursor)
install: build
	@echo "📦 Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp bin/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Binary installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "💡 Make sure $(INSTALL_DIR) is in your PATH"
	@if echo "$(MAKECMDGOALS)" | grep -q claude; then \
		echo "🔧 Configuring Claude Desktop..."; \
		$(MAKE) install-claude-config; \
	fi
	@if echo "$(MAKECMDGOALS)" | grep -q cursor; then \
		echo "🔧 Configuring Cursor..."; \
		$(MAKE) install-cursor-config; \
	fi
	@echo "✅ Installation complete!"

# Configure Claude Desktop MCP settings
install-claude-config:
	@mkdir -p "$(CLAUDE_CONFIG_DIR)"
	@if [ -f "$(CLAUDE_CONFIG_FILE)" ]; then \
		echo "📝 Updating existing Claude Desktop config..."; \
		python3 -c "import json, sys; \
			config = json.load(open('$(CLAUDE_CONFIG_FILE)')) if open('$(CLAUDE_CONFIG_FILE)').read().strip() else {'mcpServers': {}}; \
			config.setdefault('mcpServers', {})['linux-system-monitor'] = { \
				'command': '$(INSTALL_DIR)/$(BINARY_NAME)', \
				'args': [], \
				'env': {}, \
				'description': 'High-performance Linux system monitoring MCP server built with Go' \
			}; \
			json.dump(config, open('$(CLAUDE_CONFIG_FILE)', 'w'), indent=2)"; \
	else \
		echo "📄 Creating new Claude Desktop config..."; \
		echo '{ \
		  "mcpServers": { \
		    "linux-system-monitor": { \
		      "command": "$(INSTALL_DIR)/$(BINARY_NAME)", \
		      "args": [], \
		      "env": {}, \
		      "description": "High-performance Linux system monitoring MCP server built with Go" \
		    } \
		  } \
		}' > "$(CLAUDE_CONFIG_FILE)"; \
	fi
	@echo "✅ Claude Desktop configured with MCP server"

# Configure Cursor MCP settings
install-cursor-config:
	@mkdir -p "$(CURSOR_CONFIG_DIR)"
	@if [ -f "$(CURSOR_CONFIG_FILE)" ]; then \
		echo "📝 Updating existing Cursor config..."; \
		python3 -c "import json, sys; \
			config = json.load(open('$(CURSOR_CONFIG_FILE)')) if open('$(CURSOR_CONFIG_FILE)').read().strip() else {'mcpServers': {}}; \
			config.setdefault('mcpServers', {})['linux-system-monitor'] = { \
				'command': '$(INSTALL_DIR)/$(BINARY_NAME)', \
				'args': [], \
				'env': {}, \
				'description': 'High-performance Linux system monitoring MCP server built with Go' \
			}; \
			json.dump(config, open('$(CURSOR_CONFIG_FILE)', 'w'), indent=2)"; \
	else \
		echo "📄 Creating new Cursor config..."; \
		echo '{ \
		  "mcpServers": { \
		    "linux-system-monitor": { \
		      "command": "$(INSTALL_DIR)/$(BINARY_NAME)", \
		      "args": [], \
		      "env": {}, \
		      "description": "High-performance Linux system monitoring MCP server built with Go" \
		    } \
		  } \
		}' > "$(CURSOR_CONFIG_FILE)"; \
	fi
	@echo "✅ Cursor configured with MCP server"

# Dummy targets for parameters
claude:
	@# This target does nothing, it's just used as a parameter

cursor:
	@# This target does nothing, it's just used as a parameter

# Uninstall the binary and remove from Claude Desktop config
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Binary removed from $(INSTALL_DIR)"
	@if [ -f "$(CLAUDE_CONFIG_FILE)" ]; then \
		echo "🔧 Removing from Claude Desktop config..."; \
		python3 -c "import json; \
			config = json.load(open('$(CLAUDE_CONFIG_FILE)')); \
			config.get('mcpServers', {}).pop('linux-system-monitor', None); \
			json.dump(config, open('$(CLAUDE_CONFIG_FILE)', 'w'), indent=2)"; \
		echo "✅ Removed from Claude Desktop config"; \
	else \
		echo "ℹ️  No Claude Desktop config found"; \
	fi
	@if [ -f "$(CURSOR_CONFIG_FILE)" ]; then \
		echo "🔧 Removing from Cursor config..."; \
		python3 -c "import json; \
			config = json.load(open('$(CURSOR_CONFIG_FILE)')); \
			config.get('mcpServers', {}).pop('linux-system-monitor', None); \
			json.dump(config, open('$(CURSOR_CONFIG_FILE)', 'w'), indent=2)"; \
		echo "✅ Removed from Cursor config"; \
	else \
		echo "ℹ️  No Cursor config found"; \
	fi
	@echo "✅ Uninstallation complete!"

# Run the server
run:
	go run .

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Show installation status
status:
	@echo "📊 Installation Status:"
	@if [ -f "$(INSTALL_DIR)/$(BINARY_NAME)" ]; then \
		echo "✅ Binary: $(INSTALL_DIR)/$(BINARY_NAME) (installed)"; \
		$(INSTALL_DIR)/$(BINARY_NAME) --version 2>/dev/null || echo "   Version: $(VERSION)"; \
	else \
		echo "❌ Binary: Not installed"; \
	fi
	@if [ -f "$(CLAUDE_CONFIG_FILE)" ] && grep -q "linux-system-monitor" "$(CLAUDE_CONFIG_FILE)"; then \
		echo "✅ Claude Desktop: Configured"; \
	else \
		echo "❌ Claude Desktop: Not configured"; \
	fi
	@if [ -f "$(CURSOR_CONFIG_FILE)" ] && grep -q "linux-system-monitor" "$(CURSOR_CONFIG_FILE)"; then \
		echo "✅ Cursor: Configured"; \
	else \
		echo "❌ Cursor: Not configured"; \
	fi
	@if echo $$PATH | grep -q "$(INSTALL_DIR)"; then \
		echo "✅ PATH: $(INSTALL_DIR) is in PATH"; \
	else \
		echo "⚠️  PATH: $(INSTALL_DIR) is NOT in PATH"; \
		echo "   Add 'export PATH=\$$PATH:$(INSTALL_DIR)' to your shell profile"; \
	fi

# Create a release build
release: clean build-linux
	tar -czf bin/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C bin $(BINARY_NAME)-linux
	@echo "✅ Release package created: bin/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz"

.PHONY: help build build-linux install install-claude-config install-cursor-config claude cursor uninstall run test clean deps fmt lint status release
