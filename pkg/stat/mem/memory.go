package mem

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/mem"
)

type System struct {
	Total        uint64  `json:"total"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type Process struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGc      uint32 `json:"num_gc"`
}

func GetSystemMemory() *System {
	info, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("mem.VirtualMemory error: %v\n", err)
		return &System{}
	}

	return &System{
		Total:        info.Total >> 20,
		Free:         info.Free >> 20,
		UsagePercent: info.UsedPercent,
	}
}

func GetProcessMemory() *Process {
	info := &runtime.MemStats{}
	runtime.ReadMemStats(info)

	return &Process{
		Alloc:      info.Alloc >> 20,
		TotalAlloc: info.TotalAlloc >> 20,
		Sys:        info.Sys >> 20,
		NumGc:      info.NumGC,
	}
}
