// +build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Preconfig(kill bool) {
	if kill {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGSTOP,
			syscall.SIGPIPE,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			for s := range sigc {
				fmt.Println("Signal catched:", s)
				fmt.Println("For stop server, close in web")
			}
		}()
	}
}
