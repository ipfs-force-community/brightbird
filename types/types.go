package types

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type TestId string
type PrivateRegistry string
type PluginStore string

func GetString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

type Shutdown chan struct{}

func CatchSig(ctx context.Context, done Shutdown) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case s := <-c:
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				break LOOP
			case syscall.SIGHUP:
			case syscall.SIGSEGV:
			default:
				break LOOP
			}
		}
	}
	fmt.Println("receive signal")
	done <- struct{}{}
}
