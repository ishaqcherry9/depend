package cpu

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	pscpu "github.com/shirou/gopsutil/v3/cpu"
)

type cgroupCPU struct {
	frequency uint64
	quota     float64
	cores     uint64

	preSystem uint64
	preTotal  uint64
}

func newCgroupCPU() (cpu *cgroupCPU, err error) {
	cores, err := pscpu.Counts(true)
	if err != nil || cores == 0 {
		var cpus []uint64
		cpus, err = perCPUUsage()
		if err != nil {
			return nil, err
		}
		cores = len(cpus)
	}

	sets, err := cpuSets()
	if err != nil {
		return nil, err
	}
	quota := float64(len(sets))
	cq, err := cpuQuota()
	if err == nil && cq != -1 {
		var period uint64
		if period, err = cpuPeriod(); err != nil {
			return nil, err
		}
		limit := float64(cq) / float64(period)
		if limit < quota {
			quota = limit
		}
	}
	maxFreq := cpuMaxFreq()

	preSystem, err := systemCPUUsage()
	if err != nil {
		return nil, err
	}
	preTotal, err := totalCPUUsage()
	if err != nil {
		return nil, err
	}
	cpu = &cgroupCPU{
		frequency: maxFreq,
		quota:     quota,
		cores:     uint64(cores),
		preSystem: preSystem,
		preTotal:  preTotal,
	}
	return cpu, nil
}

func (cpu *cgroupCPU) Usage() (u uint64, err error) {
	var (
		total  uint64
		system uint64
	)
	total, err = totalCPUUsage()
	if err != nil {
		return 0, err
	}
	system, err = systemCPUUsage()
	if err != nil {
		return 0, err
	}
	if system != cpu.preSystem {
		u = uint64(float64((total-cpu.preTotal)*cpu.cores*1e3) / (float64(system-cpu.preSystem) * cpu.quota))
	}
	cpu.preSystem = system
	cpu.preTotal = total
	return u, err
}

func (cpu *cgroupCPU) Info() Info {
	return Info{
		Frequency: cpu.frequency,
		Quota:     cpu.quota,
	}
}

const nanoSecondsPerSecond = 1e9

var ErrNoCFSLimit = errors.New("no quota limit")

var clockTicksPerSecond = uint64(getClockTicks())

func systemCPUUsage() (usage uint64, err error) {
	var (
		line string
		f    *os.File
	)
	if f, err = os.Open("/proc/stat"); err != nil {
		return usage, err
	}
	bufReader := bufio.NewReaderSize(nil, 128)
	defer func() {
		bufReader.Reset(nil)
		_ = f.Close()
	}()
	bufReader.Reset(f)
	for err == nil {
		if line, err = bufReader.ReadString('\n'); err != nil {
			return usage, err
		}
		parts := strings.Fields(line)
		if parts[0] == "cpu" {
			if len(parts) < 8 {
				err = errors.New("bad format of cpu stats")
				return usage, err
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				var v uint64
				if v, err = strconv.ParseUint(i, 10, 64); err != nil {
					return usage, err
				}
				totalClockTicks += v
			}
			usage = (totalClockTicks * nanoSecondsPerSecond) / clockTicksPerSecond
			return usage, nil
		}
	}
	err = errors.New("bad stats format")
	return usage, err
}

func totalCPUUsage() (usage uint64, err error) {
	var cg *cgroup
	if cg, err = currentcGroup(); err != nil {
		return 0, err
	}
	return cg.CPUAcctUsage()
}

func perCPUUsage() (usage []uint64, err error) {
	var cg *cgroup
	if cg, err = currentcGroup(); err != nil {
		return nil, err
	}
	return cg.CPUAcctUsagePerCPU()
}

func cpuSets() (sets []uint64, err error) {
	var cg *cgroup
	if cg, err = currentcGroup(); err != nil {
		return nil, err
	}
	return cg.CPUSetCPUs()
}

func cpuQuota() (quota int64, err error) {
	var cg *cgroup
	if cg, err = currentcGroup(); err != nil {
		return 0, err
	}
	return cg.CPUCFSQuotaUs()
}

func cpuPeriod() (peroid uint64, err error) {
	var cg *cgroup
	if cg, err = currentcGroup(); err != nil {
		return 0, err
	}
	return cg.CPUCFSPeriodUs()
}

func cpuFreq() uint64 {
	lines, err := readLines("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		if key == "cpu MHz" || key == "clock" {
			if t, err := strconv.ParseFloat(strings.Replace(value, "MHz", "", 1), 64); err == nil {
				return uint64(t * 1000.0 * 1000.0)
			}
		}
	}
	return 0
}

func cpuMaxFreq() uint64 {
	feq := cpuFreq()
	data, err := readFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")
	if err != nil {
		return feq
	}

	cfeq, err := parseUint(data)
	if err == nil {
		feq = cfeq
	}
	return feq
}

func getClockTicks() int {

	return 100
}
