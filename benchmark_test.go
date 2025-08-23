package main

import (
	"context"
	"testing"
)

// Benchmark tests for performance measurement

func BenchmarkGetSystemInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getSystemInfo(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCPUInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getCPUInfo(ctx, false, 100)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetMemoryInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getMemoryInfo(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetDiskInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getDiskInfo(ctx, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetNetworkInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getNetworkInfo(ctx, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetProcessInfo(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getProcessInfo(ctx, 0, "", 5, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetLoadAverage(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := getLoadAverage(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSortProcesses(b *testing.B) {
	processes := []ProcessInfo{
		{PID: 100, Name: "process_c", CPUPercent: 15.5, MemoryPercent: 10.0},
		{PID: 50, Name: "process_a", CPUPercent: 25.0, MemoryPercent: 20.0},
		{PID: 75, Name: "process_b", CPUPercent: 5.5, MemoryPercent: 30.0},
		{PID: 200, Name: "process_d", CPUPercent: 35.0, MemoryPercent: 5.0},
		{PID: 150, Name: "process_e", CPUPercent: 8.0, MemoryPercent: 40.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testProcesses := make([]ProcessInfo, len(processes))
		copy(testProcesses, processes)
		sortProcessesBy(testProcesses, "cpu")
	}
}
