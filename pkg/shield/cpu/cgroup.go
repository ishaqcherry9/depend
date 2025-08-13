package cpu

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const cgroupRootDir = "/sys/fs/cgroup"

type cgroup struct {
	cgroupSet map[string]string
}

func (c *cgroup) CPUCFSQuotaUs() (int64, error) {
	data, err := readFile(path.Join(c.cgroupSet["cpu"], "cpu.cfs_quota_us"))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(data, 10, 64)
}

func (c *cgroup) CPUCFSPeriodUs() (uint64, error) {
	data, err := readFile(path.Join(c.cgroupSet["cpu"], "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}
	return parseUint(data)
}

func (c *cgroup) CPUAcctUsage() (uint64, error) {
	data, err := readFile(path.Join(c.cgroupSet["cpuacct"], "cpuacct.usage"))
	if err != nil {
		return 0, err
	}
	return parseUint(data)
}

func (c *cgroup) CPUAcctUsagePerCPU() ([]uint64, error) {
	data, err := readFile(path.Join(c.cgroupSet["cpuacct"], "cpuacct.usage_percpu"))
	if err != nil {
		return nil, err
	}
	var usage []uint64
	for _, v := range strings.Fields(data) {
		var u uint64
		if u, err = parseUint(v); err != nil {
			return nil, err
		}

		if u != 0 {
			usage = append(usage, u)
		}
	}
	return usage, nil
}

func (c *cgroup) CPUSetCPUs() ([]uint64, error) {
	data, err := readFile(path.Join(c.cgroupSet["cpuset"], "cpuset.cpus"))
	if err != nil {
		return nil, err
	}
	cpus, err := ParseUintList(data)
	if err != nil {
		return nil, err
	}
	sets := make([]uint64, 0)
	for k := range cpus {
		sets = append(sets, uint64(k))
	}
	return sets, nil
}

func currentcGroup() (*cgroup, error) {
	pid := os.Getpid()
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", pid)
	cgroupSet := make(map[string]string)
	fp, err := os.Open(cgroupFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	buf := bufio.NewReader(fp)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		col := strings.Split(strings.TrimSpace(line), ":")
		if len(col) != 3 {
			return nil, fmt.Errorf("invalid cgroup format %s", line)
		}
		dir := col[2]

		if dir != "/" {
			cgroupSet[col[1]] = path.Join(cgroupRootDir, col[1])
			if strings.Contains(col[1], ",") {
				for _, k := range strings.Split(col[1], ",") {
					cgroupSet[k] = path.Join(cgroupRootDir, k)
				}
			}
		} else {
			cgroupSet[col[1]] = path.Join(cgroupRootDir, col[1], col[2])
			if strings.Contains(col[1], ",") {
				for _, k := range strings.Split(col[1], ",") {
					cgroupSet[k] = path.Join(cgroupRootDir, k, col[2])
				}
			}
		}
	}
	return &cgroup{cgroupSet: cgroupSet}, nil
}
