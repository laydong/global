package signal

import (
	"fmt"
	"os"
	"os/signal"
)

var gs = make(map[os.Signal]func())
var sigterm = make(chan os.Signal)

func RegisterSignalFunc(f func(), signals ...os.Signal) {
	for _, item := range signals {
		gs[item] = f
	}
}

func InitSignal() {
	signal.Notify(sigterm)
	go func() {
		for s := range sigterm {
			fmt.Printf("SIGNAL %d\r\n", s)
			if f, ok := gs[s]; ok {
				f()
			}
		}
	}()
}
