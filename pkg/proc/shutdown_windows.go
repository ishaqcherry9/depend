//go:build windows

package proc

import "time"

type ShutdownConf struct{}

func AddShutdownListener(fn func()) func() {
	return fn
}

func AddWrapUpListener(fn func()) func() {
	return fn
}

func SetTimeToForceQuit(duration time.Duration) {
}

func Setup(conf ShutdownConf) {
}

func Shutdown() {
}

func WrapUp() {
}
