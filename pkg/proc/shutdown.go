//go:build linux || darwin || freebsd

package proc

import (
	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/ishaqcherry9/depend/pkg/threading"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	defaultWrapUpTime = time.Second
	defaultWaitTime   = 5500 * time.Millisecond
)

var (
	wrapUpListeners   = new(listenerManager)
	shutdownListeners = new(listenerManager)
	wrapUpTime        = defaultWrapUpTime
	waitTime          = defaultWaitTime
	shutdownLock      sync.Mutex
)

type ShutdownConf struct {
	WrapUpTime time.Duration `json:",default=1s"`

	WaitTime time.Duration `json:",default=5.5s"`
}

func AddShutdownListener(fn func()) (waitForCalled func()) {
	return shutdownListeners.addListener(fn)
}

func AddWrapUpListener(fn func()) (waitForCalled func()) {
	return wrapUpListeners.addListener(fn)
}

func SetTimeToForceQuit(duration time.Duration) {
	shutdownLock.Lock()
	defer shutdownLock.Unlock()
	waitTime = duration
}

func Setup(conf ShutdownConf) {
	shutdownLock.Lock()
	defer shutdownLock.Unlock()

	if conf.WrapUpTime > 0 {
		wrapUpTime = conf.WrapUpTime
	}
	if conf.WaitTime > 0 {
		waitTime = conf.WaitTime
	}
}

func Shutdown() {
	shutdownListeners.notifyListeners()
}

func WrapUp() {
	wrapUpListeners.notifyListeners()
}

func gracefulStop(signals chan os.Signal, sig syscall.Signal) {
	signal.Stop(signals)

	logger.Infof("Got signal %d, shutting down...", sig)
	go wrapUpListeners.notifyListeners()

	time.Sleep(wrapUpTime)
	go shutdownListeners.notifyListeners()

	shutdownLock.Lock()
	remainingTime := waitTime - wrapUpTime
	shutdownLock.Unlock()

	time.Sleep(remainingTime)
	logger.Infof("Still alive after %v, going to force kill the process...", waitTime)
	_ = syscall.Kill(syscall.Getpid(), sig)
}

type listenerManager struct {
	lock      sync.Mutex
	waitGroup sync.WaitGroup
	listeners []func()
}

func (lm *listenerManager) addListener(fn func()) (waitForCalled func()) {
	lm.waitGroup.Add(1)

	lm.lock.Lock()
	lm.listeners = append(lm.listeners, func() {
		defer lm.waitGroup.Done()
		fn()
	})
	lm.lock.Unlock()

	return func() {
		lm.waitGroup.Wait()
	}
}

func (lm *listenerManager) notifyListeners() {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	group := threading.NewRoutineGroup()
	for _, listener := range lm.listeners {
		group.RunSafe(listener)
	}
	group.Wait()

	lm.listeners = nil
}
