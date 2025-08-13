package cpu

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	interval time.Duration = time.Millisecond * 500
)

var (
	stats CPU
	usage uint64
)

type CPU interface {
	Usage() (u uint64, e error)
	Info() Info
}

func init() {
	var err error
	stats, err = newCgroupCPU()
	if err != nil {
		errStr := err.Error()
		stats, err = newPsutilCPU(interval)
		if err != nil {
			errStr += " | " + err.Error()
			fmt.Printf(`
[ERROR] cgroup cpu init and psutil cpu init are all failed, %s.
does not support getting this CPU information.
`, errStr)
			return
		}
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			u, err := stats.Usage()
			if err == nil && u != 0 {
				atomic.StoreUint64(&usage, u)
			}
		}
	}()
}

type Stat struct {
	Usage uint64
}

type Info struct {
	Frequency uint64
	Quota     float64
}

func ReadStat(stat *Stat) {
	stat.Usage = atomic.LoadUint64(&usage)
}

func GetInfo() Info {
	return stats.Info()
}
