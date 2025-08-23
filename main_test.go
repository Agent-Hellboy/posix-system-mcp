package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test utility functions
func TestTextOK(t *testing.T) {
	result := textOK("test message")
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "test message", textContent.Text)
}

func TestTextErr(t *testing.T) {
	err := assert.AnError
	result := textErr(err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Error: "+err.Error(), textContent.Text)
}

func TestSortProcessesBy(t *testing.T) {
	processes := []ProcessInfo{
		{PID: 100, Name: "process_c", CPUPercent: 15.5, MemoryPercent: 10.0},
		{PID: 50, Name: "process_a", CPUPercent: 25.0, MemoryPercent: 20.0},
		{PID: 75, Name: "process_b", CPUPercent: 5.5, MemoryPercent: 30.0},
	}

	t.Run("sort by CPU", func(t *testing.T) {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		
		sortProcessesBy(testProcesses, "cpu")
		
		assert.Equal(t, float64(25.0), testProcesses[0].CPUPercent)
		assert.Equal(t, float64(15.5), testProcesses[1].CPUPercent)
		assert.Equal(t, float64(5.5), testProcesses[2].CPUPercent)
	})

	t.Run("sort by memory", func(t *testing.T) {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		
		sortProcessesBy(testProcesses, "memory")
		
		assert.Equal(t, float32(30.0), testProcesses[0].MemoryPercent)
		assert.Equal(t, float32(20.0), testProcesses[1].MemoryPercent)
		assert.Equal(t, float32(10.0), testProcesses[2].MemoryPercent)
	})

	t.Run("sort by PID", func(t *testing.T) {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		
		sortProcessesBy(testProcesses, "pid")
		
		assert.Equal(t, int32(50), testProcesses[0].PID)
		assert.Equal(t, int32(75), testProcesses[1].PID)
		assert.Equal(t, int32(100), testProcesses[2].PID)
	})

	t.Run("sort by name", func(t *testing.T) {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		
		sortProcessesBy(testProcesses, "name")
		
		assert.Equal(t, "process_a", testProcesses[0].Name)
		assert.Equal(t, "process_b", testProcesses[1].Name)
		assert.Equal(t, "process_c", testProcesses[2].Name)
	})

	t.Run("default sort (CPU)", func(t *testing.T) {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		
		sortProcessesBy(testProcesses, "")
		
		assert.Equal(t, float64(25.0), testProcesses[0].CPUPercent)
		assert.Equal(t, float64(15.5), testProcesses[1].CPUPercent)
		assert.Equal(t, float64(5.5), testProcesses[2].CPUPercent)
	})
}

// Test system monitoring functions
func TestGetSystemInfo(t *testing.T) {
	ctx := context.Background()
	
	info, err := getSystemInfo(ctx)
	require.NoError(t, err)
	
	assert.NotEmpty(t, info.Hostname)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Platform)
	assert.NotEmpty(t, info.KernelVersion)
	assert.Greater(t, info.Uptime, uint64(0))
	assert.Greater(t, info.BootTime, uint64(0))
}

func TestGetCPUInfo(t *testing.T) {
	ctx := context.Background()
	
	t.Run("aggregate CPU info", func(t *testing.T) {
		info, err := getCPUInfo(ctx, false, 100)
		require.NoError(t, err)
		
		assert.Greater(t, len(info.Usage), 0)
		assert.Greater(t, info.Count, 0)
		assert.Greater(t, info.PhysicalCount, 0)
		assert.NotEmpty(t, info.ModelName)
	})

	t.Run("per-CPU info", func(t *testing.T) {
		info, err := getCPUInfo(ctx, true, 100)
		require.NoError(t, err)
		
		assert.Greater(t, len(info.Usage), 1) // Should have multiple CPU cores
		assert.Greater(t, info.Count, 0)
		assert.Greater(t, info.PhysicalCount, 0)
	})

	t.Run("interval clamping", func(t *testing.T) {
		// Test that intervals are clamped to valid range
		info, err := getCPUInfo(ctx, false, 50) // Too low, should be clamped to 100
		require.NoError(t, err)
		assert.Greater(t, len(info.Usage), 0)

		info, err = getCPUInfo(ctx, false, 20000) // Too high, should be clamped to 10000
		require.NoError(t, err)
		assert.Greater(t, len(info.Usage), 0)
	})
}

func TestGetMemoryInfo(t *testing.T) {
	ctx := context.Background()
	
	info, err := getMemoryInfo(ctx)
	require.NoError(t, err)
	
	assert.Greater(t, info.Total, uint64(0))
	assert.Greater(t, info.Available, uint64(0))
	assert.GreaterOrEqual(t, info.Total, info.Used)
	assert.GreaterOrEqual(t, info.UsedPercent, float64(0))
	assert.LessOrEqual(t, info.UsedPercent, float64(100))
}

func TestGetDiskInfo(t *testing.T) {
	ctx := context.Background()
	
	t.Run("all disks", func(t *testing.T) {
		result, err := getDiskInfo(ctx, "")
		require.NoError(t, err)
		
		assert.Greater(t, len(result.Disks), 0)
		for _, disk := range result.Disks {
			assert.NotEmpty(t, disk.Mountpoint)
			assert.Greater(t, disk.Total, uint64(0))
			assert.GreaterOrEqual(t, disk.Total, disk.Used)
			assert.GreaterOrEqual(t, disk.UsedPercent, float64(0))
			assert.LessOrEqual(t, disk.UsedPercent, float64(100))
		}
	})

	t.Run("specific path", func(t *testing.T) {
		result, err := getDiskInfo(ctx, "/")
		require.NoError(t, err)
		
		assert.Len(t, result.Disks, 1)
		disk := result.Disks[0]
		assert.Equal(t, "/", disk.Mountpoint)
		assert.Greater(t, disk.Total, uint64(0))
	})
}

func TestGetNetworkInfo(t *testing.T) {
	ctx := context.Background()
	
	t.Run("all interfaces", func(t *testing.T) {
		result, err := getNetworkInfo(ctx, "")
		require.NoError(t, err)
		
		assert.Greater(t, len(result.Interfaces), 0)
		for _, iface := range result.Interfaces {
			assert.NotEmpty(t, iface.Interface)
		}
	})

	t.Run("specific interface", func(t *testing.T) {
		// First get all interfaces to find a valid one
		allResult, err := getNetworkInfo(ctx, "")
		require.NoError(t, err)
		
		if len(allResult.Interfaces) > 0 {
			targetInterface := allResult.Interfaces[0].Interface
			result, err := getNetworkInfo(ctx, targetInterface)
			require.NoError(t, err)
			
			assert.Len(t, result.Interfaces, 1)
			assert.Equal(t, targetInterface, result.Interfaces[0].Interface)
		}
	})
}

func TestGetProcessInfo(t *testing.T) {
	ctx := context.Background()
	
	t.Run("default parameters", func(t *testing.T) {
		result, err := getProcessInfo(ctx, 0, "", 0, "")
		require.NoError(t, err)
		
		assert.Greater(t, len(result.Processes), 0)
		assert.LessOrEqual(t, len(result.Processes), 10) // Default limit
		assert.Equal(t, len(result.Processes), result.Count)
		
		// Should be sorted by CPU by default
		if len(result.Processes) > 1 {
			assert.GreaterOrEqual(t, result.Processes[0].CPUPercent, result.Processes[1].CPUPercent)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		result, err := getProcessInfo(ctx, 0, "", 5, "")
		require.NoError(t, err)
		
		assert.LessOrEqual(t, len(result.Processes), 5)
	})

	t.Run("with name filter", func(t *testing.T) {
		// Try to find a common process name
		allResult, err := getProcessInfo(ctx, 0, "", 50, "")
		require.NoError(t, err)
		
		if len(allResult.Processes) > 0 {
			// Use part of the first process name for filtering
			processName := allResult.Processes[0].Name
			if len(processName) > 3 {
				filterName := processName[:3] // Use first 3 characters
				result, err := getProcessInfo(ctx, 0, filterName, 10, "")
				require.NoError(t, err)
				
				// All returned processes should contain the filter string
				for _, proc := range result.Processes {
					assert.Contains(t, strings.ToLower(proc.Name), strings.ToLower(filterName))
				}
			}
		}
	})

	t.Run("sort by memory", func(t *testing.T) {
		result, err := getProcessInfo(ctx, 0, "", 5, "memory")
		require.NoError(t, err)
		
		if len(result.Processes) > 1 {
			assert.GreaterOrEqual(t, result.Processes[0].MemoryPercent, result.Processes[1].MemoryPercent)
		}
	})

	t.Run("specific PID", func(t *testing.T) {
		// Use PID 1 which should always exist on Linux systems
		result, err := getProcessInfo(ctx, 1, "", 0, "")
		require.NoError(t, err)
		
		assert.Len(t, result.Processes, 1)
		assert.Equal(t, int32(1), result.Processes[0].PID)
	})

	t.Run("limit bounds", func(t *testing.T) {
		// Test limit clamping
		result, err := getProcessInfo(ctx, 0, "", 300, "") // Too high, should be clamped to 200
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result.Processes), 200)
		
		result, err = getProcessInfo(ctx, 0, "", -5, "") // Negative, should use default of 10
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result.Processes), 10)
	})
}

func TestGetLoadAverage(t *testing.T) {
	ctx := context.Background()
	
	result, err := getLoadAverage(ctx)
	require.NoError(t, err)
	
	assert.GreaterOrEqual(t, result.Load1, float64(0))
	assert.GreaterOrEqual(t, result.Load5, float64(0))
	assert.GreaterOrEqual(t, result.Load15, float64(0))
}

// Integration tests
func TestServerRegistration(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)

	// This should not panic
	require.NotPanics(t, func() {
		registerTools(server)
	})
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "linux-system-mcp", ServerName)
	assert.Equal(t, "1.0.0", ServerVersion)
	assert.NotEmpty(t, Version)
}

// Test data structure validation
func TestDataStructures(t *testing.T) {
	t.Run("SystemInfo", func(t *testing.T) {
		info := SystemInfo{
			Hostname:        "test-host",
			OS:              "linux",
			Platform:        "ubuntu",
			PlatformFamily:  "debian",
			PlatformVersion: "20.04",
			KernelVersion:   "5.4.0",
			KernelArch:      "x86_64",
			Uptime:          3600,
			BootTime:        1234567890,
			Procs:           100,
			HostID:          "test-host-id",
		}
		
		assert.Equal(t, "test-host", info.Hostname)
		assert.Equal(t, "linux", info.OS)
		assert.Equal(t, uint64(3600), info.Uptime)
	})

	t.Run("CPUInfo", func(t *testing.T) {
		info := CPUInfo{
			Usage:         []float64{25.5, 30.0},
			Count:         4,
			PhysicalCount: 2,
			ModelName:     "Test CPU",
			Family:        "Test Family",
			Speed:         2400.0,
			CacheSize:     8192,
			Flags:         []string{"flag1", "flag2"},
		}
		
		assert.Len(t, info.Usage, 2)
		assert.Equal(t, 4, info.Count)
		assert.Equal(t, "Test CPU", info.ModelName)
	})

	t.Run("ProcessInfo", func(t *testing.T) {
		info := ProcessInfo{
			PID:           1234,
			Name:          "test-process",
			Status:        "running",
			CPUPercent:    15.5,
			MemoryRSS:     1024000,
			MemoryVMS:     2048000,
			MemoryPercent: 5.5,
			CreateTime:    1234567890,
			NumThreads:    4,
			Username:      "testuser",
			Cmdline:       []string{"test-process", "--arg"},
		}
		
		assert.Equal(t, int32(1234), info.PID)
		assert.Equal(t, "test-process", info.Name)
		assert.Equal(t, float64(15.5), info.CPUPercent)
		assert.Len(t, info.Cmdline, 2)
	})
}
