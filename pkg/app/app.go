package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/ishaqcherry9/depend/pkg/prof"
)

type IServer interface {
	Start() error
	Stop() error
	String() string
}

type Close func() error

type App struct {
	servers []IServer
	closes  []Close
}

func New(servers []IServer, closes []Close) *App {
	return &App{
		servers: servers,
		closes:  closes,
	}
}

func (a *App) Run() {

	eg, ctx := errgroup.WithContext(context.Background())

	for _, server := range a.servers {
		s := server
		eg.Go(func() error {
			fmt.Println(s.String())
			return s.Start()
		})
	}

	eg.Go(func() error {
		return a.watch(ctx)
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

func (a *App) watch(ctx context.Context) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGTRAP)
	profile := prof.NewProfile()

	for {
		select {
		case <-ctx.Done():
			_ = a.stop()
			return ctx.Err()

		case sigType := <-sig:
			fmt.Printf("received system notification signal: %s\n", sigType.String())
			switch sigType {
			case syscall.SIGTRAP:
				profile.StartOrStop()
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP:
				if err := a.stop(); err != nil {
					return err
				}
				fmt.Println("stop app successfully")
				return nil
			}
		}
	}
}

func (a *App) stop() error {
	for _, closeFn := range a.closes {
		if err := closeFn(); err != nil {
			return err
		}
	}
	return nil
}
