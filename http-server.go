package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// HTTPServer wraps the MCP server for HTTP transport
type HTTPServer struct {
	config *ServerConfig
}

// ServerConfig holds configuration from Smithery
type ServerConfig struct {
	RefreshInterval int  `json:"refreshInterval"`
	EnableDebug     bool `json:"enableDebug"`
}

// MCPRequest represents an incoming MCP JSON-RPC request
type MCPRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an outgoing MCP JSON-RPC response
type MCPResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// NewHTTPServer creates a new HTTP server wrapper
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		config: &ServerConfig{RefreshInterval: 1000, EnableDebug: false},
	}
}

// ServeHTTP handles HTTP requests with CORS
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers as required by Smithery
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "mcp-session-id, mcp-protocol-version")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse configuration from URL if present
	if configParam := r.URL.Query().Get("config"); configParam != "" {
		if err := h.parseConfig(configParam); err != nil && h.config.EnableDebug {
			log.Printf("Failed to parse config: %v", err)
		}
	}

	// Handle MCP requests
	h.handleMCPRequest(w, r)
}

// handleMCPRequest processes MCP JSON-RPC requests
func (h *HTTPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if h.config.EnableDebug {
		log.Printf("MCP Request: %s", string(body))
	}

	ctx := context.Background()
	var result interface{}
	var mcpError error

	// Handle MCP methods
	switch req.Method {
	case "initialize":
		result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    ServerName,
				"version": Version,
			},
		}

	case "tools/list":
		result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "get_system_info",
					"description": "Get comprehensive system information including hostname, OS, platform, uptime, etc.",
				},
				{
					"name":        "get_cpu_info",
					"description": "Get detailed CPU usage statistics and information",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"per_cpu": map[string]interface{}{
								"type":        "boolean",
								"description": "Get per-CPU usage if true",
							},
							"interval_ms": map[string]interface{}{
								"type":        "integer",
								"description": "Sampling window in ms (100..10000)",
							},
						},
					},
				},
				{
					"name":        "get_memory_info",
					"description": "Get memory usage information including RAM and swap",
				},
				{
					"name":        "get_disk_info",
					"description": "Get disk usage information for all partitions or a specific path",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"path": map[string]interface{}{
								"type":        "string",
								"description": "Specific path to check; if empty, all mounts",
							},
						},
					},
				},
				{
					"name":        "get_network_info",
					"description": "Get network interface statistics and information",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"interface": map[string]interface{}{
								"type":        "string",
								"description": "Specific interface to include",
							},
						},
					},
				},
				{
					"name":        "get_process_info",
					"description": "Get information about running processes with filtering and sorting options",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"pid": map[string]interface{}{
								"type":        "integer",
								"description": "Specific PID",
							},
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Filter by name substring",
							},
							"limit": map[string]interface{}{
								"type":        "integer",
								"description": "Max results (1..200, default 10)",
							},
							"sort_by": map[string]interface{}{
								"type":        "string",
								"description": "Sort by: cpu|memory|pid|name",
							},
						},
					},
				},
				{
					"name":        "get_load_average",
					"description": "Get system load average (1, 5, and 15 minute averages)",
				},
			},
		}

	case "tools/call":
		result, mcpError = h.handleToolCall(ctx, req.Params)

	default:
		mcpError = fmt.Errorf("unknown method: %s", req.Method)
	}

	// Build response
	response := MCPResponse{
		JsonRPC: "2.0",
		ID:      req.ID,
	}

	if mcpError != nil {
		response.Error = map[string]interface{}{
			"code":    -32601,
			"message": mcpError.Error(),
		}
	} else {
		response.Result = result
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleToolCall processes tool call requests
func (h *HTTPServer) handleToolCall(ctx context.Context, params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params format")
	}

	toolName, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	var arguments map[string]interface{}
	if args, exists := paramsMap["arguments"]; exists {
		if argsMap, ok := args.(map[string]interface{}); ok {
			arguments = argsMap
		}
	}

	if h.config.EnableDebug {
		log.Printf("Tool call: %s with args: %+v", toolName, arguments)
	}

	// Call the appropriate tool function
	switch toolName {
	case "get_system_info":
		result, err := getSystemInfo(ctx)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_cpu_info":
		perCPU := false
		intervalMs := 1000
		if arguments != nil {
			if val, ok := arguments["per_cpu"].(bool); ok {
				perCPU = val
			}
			if val, ok := arguments["interval_ms"].(float64); ok {
				intervalMs = int(val)
			}
		}
		result, err := getCPUInfo(ctx, perCPU, intervalMs)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_memory_info":
		result, err := getMemoryInfo(ctx)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_disk_info":
		path := ""
		if arguments != nil {
			if val, ok := arguments["path"].(string); ok {
				path = val
			}
		}
		result, err := getDiskInfo(ctx, path)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_network_info":
		iface := ""
		if arguments != nil {
			if val, ok := arguments["interface"].(string); ok {
				iface = val
			}
		}
		result, err := getNetworkInfo(ctx, iface)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_process_info":
		var pid int32 = 0
		name := ""
		limit := 10
		sortBy := ""
		if arguments != nil {
			if val, ok := arguments["pid"].(float64); ok {
				pid = int32(val)
			}
			if val, ok := arguments["name"].(string); ok {
				name = val
			}
			if val, ok := arguments["limit"].(float64); ok {
				limit = int(val)
			}
			if val, ok := arguments["sort_by"].(string); ok {
				sortBy = val
			}
		}
		result, err := getProcessInfo(ctx, pid, name, limit, sortBy)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	case "get_load_average":
		result, err := getLoadAverage(ctx)
		return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%+v", result)}}}, err

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// parseConfig decodes and parses base64-encoded configuration
func (h *HTTPServer) parseConfig(configParam string) error {
	decoded, err := base64.StdEncoding.DecodeString(configParam)
	if err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(decoded, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if config.RefreshInterval < 100 || config.RefreshInterval > 10000 {
		config.RefreshInterval = 1000 // Default fallback
	}

	h.config = &config
	
	if h.config.EnableDebug {
		log.Printf("Configuration updated: %+v", h.config)
	}

	return nil
}

// StartHTTPServer starts the HTTP server for Smithery deployment
func StartHTTPServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Smithery default
	}

	httpServer := NewHTTPServer()
	
	http.Handle("/mcp", httpServer)
	
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": ServerName,
			"version": Version,
		})
	})

	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("MCP endpoint: /mcp")
	log.Printf("Health check: /health")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
