package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
)

var (
	LOG     *log.Logger
	CFG     *Config
	Beeline *Uplink
	TTK     *Uplink
)

func initInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(c chan os.Signal) {
		for range c {
			LOG.Println("-- stop --")
			os.Exit(137)
		}
	}(c)
}

func init() {
	runtime.GOMAXPROCS(1)
	LOG = initLogging()
	initInterrupt()
}
