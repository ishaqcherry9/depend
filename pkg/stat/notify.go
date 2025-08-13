//go:build linux || darwin
// +build linux darwin

package stat

import (
	"fmt"
	"syscall"
)

func init() {
	go func() {
		for {
			select {
			case <-notifyCh:
				err := syscall.Kill(syscall.Getpid(), syscall.SIGTRAP)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}
