package cpu

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

// System 系统CPU信息
type System struct {
	UsagePercent float64   `json:"usage_percent"` // cpu使用率，单位(%)
	CPUInfo      []CPUInfo `json:"cpu_info"`
}

// CPUInfo cpu info
type CPUInfo struct { //nolint
	ModelName string  `json:"modelName"`
	Cores     int32   `json:"cores"`
	Frequency float64 `json:"frequency"` // cpu频率，单位(Mhz)
}

// Process 进程的cpu使用率
type Process struct {
	UsagePercent float64 `json:"usage_percent"` // cpu使用率，单位(%)

	RSS uint64 `json:"rss"` // 使用物理内存，单位(M)
	VMS uint64 `json:"vms"` // 使用虚拟内存，单位(M)
}

// GetSystemCPU get system cpu info
func GetSystemCPU() *System {
	sysUsagePercent := 0.0
	vs, err := cpu.Percent(time.Millisecond*10, false)
	if err != nil {
		fmt.Printf("cpu.Percent error, %v\n", err)
	}
	if len(vs) == 1 {
		sysUsagePercent = vs[0]
	}

	var cpuInfos []CPUInfo
	cpus, err := cpu.Info()
	if err != nil {
		fmt.Printf("cpu.Info error, %v\n", err)
	} else {
		for _, v := range cpus {
			cpuInfos = append(cpuInfos, CPUInfo{
				ModelName: v.ModelName,
				Cores:     v.Cores,
				Frequency: v.Mhz,
			})
		}
	}

	return &System{
		UsagePercent: floatRound(sysUsagePercent, 1),
		CPUInfo:      cpuInfos,
	}
}

// GetProcess get current process info
func GetProcess() *Process {
	proc := &Process{}

	currentPid := os.Getpid()
	p, err := process.NewProcess(int32(currentPid))
	if err != nil {
		fmt.Printf("process.NewProcess error, %v\n", err)
		return proc
	}

	percent, err := p.CPUPercent()
	if err != nil {
		fmt.Printf("p.CPUPercent error, %v\n", err)
		return proc
	}
	proc.UsagePercent = floatRound(percent, 1)

	mInfo, _ := p.MemoryInfo()
	proc.RSS = mInfo.RSS >> 20
	proc.VMS = mInfo.VMS >> 20

	return proc
}

func floatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}
