package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	ServerName    = "posix-system-mcp"
	ServerVersion = "1.0.0"
)

var Version = "1.0.0" // Can be overridden at build time

// --- Data types ---

type SystemInfo struct {
	Hostname             string            `json:"hostname"`
	OS                   string            `json:"os"`
	Platform             string            `json:"platform"`
	PlatformFamily       string            `json:"platform_family"`
	PlatformVersion      string            `json:"platform_version"`
	KernelVersion        string            `json:"kernel_version"`
	KernelArch           string            `json:"kernel_arch"`
	Uptime               uint64            `json:"uptime_seconds"`
	BootTime             uint64            `json:"boot_time"`
	Procs                uint64            `json:"processes"`
	HostID               string            `json:"host_id"`
	VirtualizationSystem string            `json:"virtualization_system,omitempty"`
	VirtualizationRole   string            `json:"virtualization_role,omitempty"`
	Temperature          []TemperatureStat `json:"temperature,omitempty"`
}

type TemperatureStat struct {
	SensorKey   string  `json:"sensor_key"`
	Temperature float64 `json:"temperature"`
}

type CPUInfo struct {
	Usage         []float64 `json:"usage_percent"`
	Count         int       `json:"logical_count"`
	PhysicalCount int       `json:"physical_count"`
	ModelName     string    `json:"model_name"`
	Family        string    `json:"family"`
	Speed         float64   `json:"speed_mhz"`
	CacheSize     int32     `json:"cache_size"`
	Flags         []string  `json:"flags,omitempty"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total_bytes"`
	Available   uint64  `json:"available_bytes"`
	Used        uint64  `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
	Free        uint64  `json:"free_bytes"`
	Buffers     uint64  `json:"buffers_bytes"`
	Cached      uint64  `json:"cached_bytes"`
	SwapTotal   uint64  `json:"swap_total_bytes"`
	SwapUsed    uint64  `json:"swap_used_bytes"`
	SwapFree    uint64  `json:"swap_free_bytes"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total_bytes"`
	Free        uint64  `json:"free_bytes"`
	Used        uint64  `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
	InodesTotal uint64  `json:"inodes_total"`
	InodesUsed  uint64  `json:"inodes_used"`
	InodesFree  uint64  `json:"inodes_free"`
}

type NetworkInfo struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Errin       uint64 `json:"errors_in"`
	Errout      uint64 `json:"errors_out"`
	Dropin      uint64 `json:"drops_in"`
	Dropout     uint64 `json:"drops_out"`
}

type ProcessInfo struct {
	PID           int32    `json:"pid"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryRSS     uint64   `json:"memory_rss_bytes"`
	MemoryVMS     uint64   `json:"memory_vms_bytes"`
	MemoryPercent float32  `json:"memory_percent"`
	CreateTime    int64    `json:"create_time"`
	NumThreads    int32    `json:"num_threads"`
	Username      string   `json:"username,omitempty"`
	Cmdline       []string `json:"cmdline,omitempty"`
}

type DiskInfoResult struct {
	Disks []DiskInfo `json:"disks"`
}

type NetworkInfoResult struct {
	Interfaces []NetworkInfo `json:"interfaces"`
}

type ProcessInfoResult struct {
	Processes []ProcessInfo `json:"processes"`
	Count     int           `json:"count"`
}

type LoadAvgResult struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// --- Tool arg structs (no jsonschema tags) ---

type SystemInfoArgs struct{}

type CPUInfoArgs struct {
	PerCPU     bool `json:"per_cpu,omitempty"`     // get per-CPU usage if true
	IntervalMs int  `json:"interval_ms,omitempty"` // sampling window in ms (100..10000), default 1000
}

type MemoryInfoArgs struct{}

type DiskInfoArgs struct {
	Path string `json:"path,omitempty"` // specific path to check; if empty, all mounts
}

type NetworkInfoArgs struct {
	Interface string `json:"interface,omitempty"` // specific interface to include
}

type ProcessInfoArgs struct {
	PID    int32  `json:"pid,omitempty"`     // specific PID
	Name   string `json:"name,omitempty"`    // filter by name substring
	Limit  int    `json:"limit,omitempty"`   // max results (1..200, default 10)
	SortBy string `json:"sort_by,omitempty"` // cpu|memory|pid|name
}

type LoadAverageArgs struct{}

// --- Main ---

func main() {
	fmt.Fprintf(os.Stderr, "Starting %s server...\n", ServerName)

	// --version / -v
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("%s version %s\n", ServerName, Version)
		os.Exit(0)
	}

	// --http flag for HTTP mode
	if len(os.Args) > 1 && os.Args[1] == "--http" {
		fmt.Fprintf(os.Stderr, "Starting HTTP server for Smithery deployment...\n")
		StartHTTPServer()
		return
	}

	// Default STDIO mode
	fmt.Fprintf(os.Stderr, "Starting STDIO transport...\n")
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)

	registerTools(server)

	fmt.Fprintf(os.Stderr, "Server created, starting transport...\n")
	transport := &mcp.StdioTransport{}

	fmt.Fprintf(os.Stderr, "Running server...\n")
	if err := server.Run(context.Background(), transport); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		log.Fatalf("Server error: %v", err)
	}
}

// --- Tool registration ---

func registerTools(server *mcp.Server) {
	// System info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_system_info",
		Description: "Get comprehensive system information including hostname, OS, platform, uptime, etc.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ SystemInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getSystemInfo(ctx)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("System information retrieved"), out, nil
	})

	// CPU info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_cpu_info",
		Description: "Get detailed CPU usage statistics and information",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a CPUInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getCPUInfo(ctx, a.PerCPU, a.IntervalMs)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("CPU information retrieved"), out, nil
	})

	// Memory info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_memory_info",
		Description: "Get memory usage information including RAM and swap",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ MemoryInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getMemoryInfo(ctx)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("Memory information retrieved"), out, nil
	})

	// Disk info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_disk_info",
		Description: "Get disk usage information for all partitions or a specific path",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a DiskInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getDiskInfo(ctx, a.Path)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("Disk information retrieved"), out, nil
	})

	// Network info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_network_info",
		Description: "Get network interface statistics and information",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a NetworkInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getNetworkInfo(ctx, a.Interface)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("Network information retrieved"), out, nil
	})

	// Process info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_process_info",
		Description: "Get information about running processes with filtering and sorting options",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a ProcessInfoArgs) (*mcp.CallToolResult, any, error) {
		out, err := getProcessInfo(ctx, a.PID, a.Name, a.Limit, a.SortBy)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("Process information retrieved"), out, nil
	})

	// Load average
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_load_average",
		Description: "Get system load average (1, 5, and 15 minute averages)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ LoadAverageArgs) (*mcp.CallToolResult, any, error) {
		out, err := getLoadAverage(ctx)
		if err != nil {
			return textErr(err), nil, err
		}
		return textOK("Load average retrieved"), out, nil
	})
}

func textOK(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}
func textErr(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + err.Error()}}}
}

// --- Implementations ---

func getSystemInfo(ctx context.Context) (SystemInfo, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("failed to get host info: %w", err)
	}

	temps, _ := host.SensorsTemperatures() // not ctx-aware in gopsutil
	tempStats := make([]TemperatureStat, len(temps))
	for i, t := range temps {
		tempStats[i] = TemperatureStat{SensorKey: t.SensorKey, Temperature: t.Temperature}
	}

	return SystemInfo{
		Hostname:             hostInfo.Hostname,
		OS:                   hostInfo.OS,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		KernelVersion:        hostInfo.KernelVersion,
		KernelArch:           hostInfo.KernelArch,
		Uptime:               hostInfo.Uptime,
		BootTime:             hostInfo.BootTime,
		Procs:                hostInfo.Procs,
		HostID:               hostInfo.HostID,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,
		Temperature:          tempStats,
	}, nil
}

func getCPUInfo(ctx context.Context, perCPU bool, intervalMs int) (CPUInfo, error) {
	interval := time.Second
	if intervalMs > 0 {
		if intervalMs < 100 {
			intervalMs = 100
		}
		if intervalMs > 10000 {
			intervalMs = 10000
		}
		interval = time.Duration(intervalMs) * time.Millisecond
	}

	var usage []float64
	var err error
	if perCPU {
		usage, err = cpu.PercentWithContext(ctx, interval, true)
	} else {
		usage, err = cpu.PercentWithContext(ctx, interval, false)
	}
	if err != nil {
		return CPUInfo{}, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	info, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("failed to get CPU info: %w", err)
	}

	logicalCount, err := cpu.Counts(true)
	if err != nil {
		logicalCount = runtime.NumCPU()
	}
	physicalCount, err := cpu.Counts(false)
	if err != nil {
		physicalCount = logicalCount
	}

	var modelName, family string
	var speed float64
	var cacheSize int32
	var flags []string
	if len(info) > 0 {
		modelName = info[0].ModelName
		family = info[0].Family
		speed = info[0].Mhz
		cacheSize = info[0].CacheSize
		flags = info[0].Flags
	}

	return CPUInfo{
		Usage:         usage,
		Count:         logicalCount,
		PhysicalCount: physicalCount,
		ModelName:     modelName,
		Family:        family,
		Speed:         speed,
		CacheSize:     cacheSize,
		Flags:         flags,
	}, nil
}

func getMemoryInfo(ctx context.Context) (MemoryInfo, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("failed to get virtual memory: %w", err)
	}
	sw, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("failed to get swap memory: %w", err)
	}
	return MemoryInfo{
		Total:       vm.Total,
		Available:   vm.Available,
		Used:        vm.Used,
		UsedPercent: vm.UsedPercent,
		Free:        vm.Free,
		Buffers:     vm.Buffers,
		Cached:      vm.Cached,
		SwapTotal:   sw.Total,
		SwapUsed:    sw.Used,
		SwapFree:    sw.Free,
	}, nil
}

func getDiskInfo(ctx context.Context, path string) (DiskInfoResult, error) {
	var disks []DiskInfo
	if path != "" {
		u, err := disk.UsageWithContext(ctx, path)
		if err != nil {
			return DiskInfoResult{}, fmt.Errorf("failed to get disk usage for %s: %w", path, err)
		}
		disks = []DiskInfo{{
			Device:      "N/A",
			Mountpoint:  path,
			Fstype:      u.Fstype,
			Total:       u.Total,
			Free:        u.Free,
			Used:        u.Used,
			UsedPercent: u.UsedPercent,
			InodesTotal: u.InodesTotal,
			InodesUsed:  u.InodesUsed,
			InodesFree:  u.InodesFree,
		}}
	} else {
		parts, err := disk.PartitionsWithContext(ctx, false)
		if err != nil {
			return DiskInfoResult{}, fmt.Errorf("failed to get disk partitions: %w", err)
		}
		for _, p := range parts {
			u, err := disk.UsageWithContext(ctx, p.Mountpoint)
			if err != nil {
				continue
			}
			disks = append(disks, DiskInfo{
				Device:      p.Device,
				Mountpoint:  p.Mountpoint,
				Fstype:      p.Fstype,
				Total:       u.Total,
				Free:        u.Free,
				Used:        u.Used,
				UsedPercent: u.UsedPercent,
				InodesTotal: u.InodesTotal,
				InodesUsed:  u.InodesUsed,
				InodesFree:  u.InodesFree,
			})
		}
	}
	return DiskInfoResult{Disks: disks}, nil
}

func getNetworkInfo(ctx context.Context, iface string) (NetworkInfoResult, error) {
	stats, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return NetworkInfoResult{}, fmt.Errorf("failed to get network stats: %w", err)
	}
	var out []NetworkInfo
	for _, s := range stats {
		if iface != "" && s.Name != iface {
			continue
		}
		out = append(out, NetworkInfo{
			Interface:   s.Name,
			BytesSent:   s.BytesSent,
			BytesRecv:   s.BytesRecv,
			PacketsSent: s.PacketsSent,
			PacketsRecv: s.PacketsRecv,
			Errin:       s.Errin,
			Errout:      s.Errout,
			Dropin:      s.Dropin,
			Dropout:     s.Dropout,
		})
	}
	return NetworkInfoResult{Interfaces: out}, nil
}

func getProcessInfo(ctx context.Context, pid int32, name string, limit int, sortBy string) (ProcessInfoResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	var list []ProcessInfo
	if pid > 0 {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return ProcessInfoResult{}, fmt.Errorf("failed to get process %d: %w", pid, err)
		}
		info, err := getProcessDetails(ctx, proc)
		if err != nil {
			return ProcessInfoResult{}, fmt.Errorf("failed to get process details: %w", err)
		}
		list = []ProcessInfo{info}
	} else {
		procs, err := process.ProcessesWithContext(ctx)
		if err != nil {
			return ProcessInfoResult{}, fmt.Errorf("failed to list processes: %w", err)
		}
		for _, p := range procs {
			info, err := getProcessDetails(ctx, p)
			if err != nil {
				continue
			}
			if name != "" && !strings.Contains(strings.ToLower(info.Name), strings.ToLower(name)) {
				continue
			}
			list = append(list, info)
			if len(list) >= limit*2 { // gather extra for better sorting
				break
			}
		}
		sortProcessesBy(list, sortBy)
		if len(list) > limit {
			list = list[:limit]
		}
	}

	return ProcessInfoResult{Processes: list, Count: len(list)}, nil
}

func getProcessDetails(ctx context.Context, proc *process.Process) (ProcessInfo, error) {
	name, _ := proc.NameWithContext(ctx)
	statusSlice, _ := proc.StatusWithContext(ctx)
	cpuPercent, _ := proc.CPUPercentWithContext(ctx)
	memInfo, _ := proc.MemoryInfoWithContext(ctx)
	memPercent, _ := proc.MemoryPercentWithContext(ctx)
	createTime, _ := proc.CreateTimeWithContext(ctx)
	numThreads, _ := proc.NumThreadsWithContext(ctx)
	username, _ := proc.UsernameWithContext(ctx)
	cmdlineStr, _ := proc.CmdlineWithContext(ctx)

	var memoryRSS, memoryVMS uint64
	if memInfo != nil {
		memoryRSS = memInfo.RSS
		memoryVMS = memInfo.VMS
	}

	var statusStr string
	if len(statusSlice) > 0 {
		statusStr = statusSlice[0]
	}

	var cmdlineSlice []string
	if cmdlineStr != "" {
		cmdlineSlice = strings.Fields(cmdlineStr)
	}

	return ProcessInfo{
		PID:           proc.Pid,
		Name:          name,
		Status:        statusStr,
		CPUPercent:    cpuPercent,
		MemoryRSS:     memoryRSS,
		MemoryVMS:     memoryVMS,
		MemoryPercent: memPercent,
		CreateTime:    createTime,
		NumThreads:    numThreads,
		Username:      username,
		Cmdline:       cmdlineSlice,
	}, nil
}

func sortProcessesBy(processes []ProcessInfo, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "memory":
		sort.SliceStable(processes, func(i, j int) bool { return processes[i].MemoryPercent > processes[j].MemoryPercent })
	case "pid":
		sort.SliceStable(processes, func(i, j int) bool { return processes[i].PID < processes[j].PID })
	case "name":
		sort.SliceStable(processes, func(i, j int) bool { return processes[i].Name < processes[j].Name })
	default: // "cpu"
		sort.SliceStable(processes, func(i, j int) bool { return processes[i].CPUPercent > processes[j].CPUPercent })
	}
}

func getLoadAverage(ctx context.Context) (LoadAvgResult, error) {
	l, err := load.AvgWithContext(ctx)
	if err != nil {
		return LoadAvgResult{}, fmt.Errorf("failed to get load average: %w", err)
	}
	return LoadAvgResult{Load1: l.Load1, Load5: l.Load5, Load15: l.Load15}, nil
}
