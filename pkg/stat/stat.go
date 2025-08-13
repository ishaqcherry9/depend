package stat

import (
	"math"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/ishaqcherry9/depend/pkg/stat/cpu"
	"github.com/ishaqcherry9/depend/pkg/stat/mem"
)

var (
	printInfoInterval = time.Minute
	zapLog, _         = zap.NewProduction()

	notifyCh = make(chan struct{})
)

type Option func(*options)

type options struct {
	enableAlarm bool
	zapFields   []zap.Field
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithPrintInterval(d time.Duration) Option {
	return func(o *options) {
		if d < time.Second {
			return
		}
		printInfoInterval = d
	}
}

func WithLog(l *zap.Logger) Option {
	return func(o *options) {
		if l == nil {
			return
		}
		zapLog = l
	}
}

func WithPrintField(fields ...zap.Field) Option {
	return func(o *options) {
		o.zapFields = fields
	}
}

func WithAlarm(opts ...AlarmOption) Option {
	return func(o *options) {
		if runtime.GOOS == "windows" {
			return
		}
		ao := &alarmOptions{}
		ao.apply(opts...)
		o.enableAlarm = true
	}
}

func Init(opts ...Option) {
	o := &options{}
	o.apply(opts...)

	go func() {
		printTick := time.NewTicker(printInfoInterval)
		defer printTick.Stop()
		sg := newStatGroup()

		for {
			select {
			case <-printTick.C:
				data := printUsageInfo(o.zapFields...)
				if o.enableAlarm {
					if sg.check(data) {
						sendSystemSignForLinux()
					}
				}
			}
		}
	}()
}

func sendSystemSignForLinux() {
	select {
	case notifyCh <- struct{}{}:
	default:
	}
}

func printUsageInfo(fields ...zap.Field) *statData {
	defer func() { _ = recover() }()

	mSys := mem.GetSystemMemory()
	mProc := mem.GetProcessMemory()
	cSys := cpu.GetSystemCPU()
	cProc := cpu.GetProcess()

	var cors int32
	for _, ci := range cSys.CPUInfo {
		cors += ci.Cores
	}

	sys := system{
		CPUUsage: cSys.UsagePercent,
		CPUCores: cors,
		MemTotal: mSys.Total,
		MemFree:  mSys.Free,
		MemUsage: float64(int(math.Round(mSys.UsagePercent))),
	}
	proc := process{
		CPUUsage:   cProc.UsagePercent,
		RSS:        cProc.RSS,
		VMS:        cProc.VMS,
		Alloc:      mProc.Alloc,
		TotalAlloc: mProc.TotalAlloc,
		Sys:        mProc.Sys,
		NumGc:      mProc.NumGc,
		Goroutines: runtime.NumGoroutine(),
	}

	fields = append(fields, zap.Any("system", sys), zap.Any("process", proc))
	zapLog.Info("statistics", fields...)

	return &statData{
		sys:  sys,
		proc: proc,
	}
}

type system struct {
	CPUUsage float64 `json:"cpu_usage"`
	CPUCores int32   `json:"cpu_cores"`
	MemTotal uint64  `json:"mem_total"`
	MemFree  uint64  `json:"mem_free"`
	MemUsage float64 `json:"mem_usage"`
}

type process struct {
	CPUUsage   float64 `json:"cpu_usage"`
	RSS        uint64  `json:"rss"`
	VMS        uint64  `json:"vms"`
	Alloc      uint64  `json:"alloc"`
	TotalAlloc uint64  `json:"total_alloc"`
	Sys        uint64  `json:"sys"`
	NumGc      uint32  `json:"num_gc"`
	Goroutines int     `json:"goroutines"`
}

type statData struct {
	sys  system
	proc process
}
