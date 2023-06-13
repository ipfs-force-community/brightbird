package types

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

var ErrT = reflect.TypeOf((*error)(nil)).Elem()
var CtxT = reflect.TypeOf((*context.Context)(nil)).Elem()
var NilVal = reflect.ValueOf(nil)
var NilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

type TestId string
type PrivateRegistry string
type PluginStore string

type BootstrapPeers []string

func PtrString(str string) *string {
	return &str
}

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
