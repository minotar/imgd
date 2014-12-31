package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

type SignalHandler struct {
	signalChannel chan os.Signal
	stopChannel   chan int
}

func (s *SignalHandler) run() {
	run := true
	for run {
		select {
		case sig := <-s.signalChannel:
			s.handleSignal(sig)
		case <-s.stopChannel:
			signal.Stop(s.signalChannel)
			run = false
		}
	}
}

func (s *SignalHandler) handleSignal(signal os.Signal) {
	switch signal {
	case syscall.SIGUSR1:
		tf, err := ioutil.TempFile("", "block")
		if err != nil {
			log.Error(err.Error())
			break
		}

		if err := pprof.Lookup("block").WriteTo(tf, 1); err != nil {
			log.Error(err.Error())
			tf.Close()
			os.Remove(tf.Name())
			break
		}

		if err := tf.Close(); err != nil {
			log.Error(err.Error())
			break
		}
		log.Info("Dumped block pprof to " + tf.Name())
	case syscall.SIGUSR2:
		tf, err := ioutil.TempFile("", "goroutine")
		if err != nil {
			log.Error(err.Error())
			break
		}

		if err := pprof.Lookup("goroutine").WriteTo(tf, 2); err != nil {
			log.Error(err.Error())
			tf.Close()
			os.Remove(tf.Name())
			break
		}

		if err := tf.Close(); err != nil {
			log.Error(err.Error())
			break
		}
		log.Info("Dumped goroutine pprof to " + tf.Name())
	}
}

func (s *SignalHandler) Stop() {
	s.stopChannel <- 1
}

func MakeSignalHandler() *SignalHandler {
	s := new(SignalHandler)
	s.stopChannel = make(chan int)
	s.signalChannel = make(chan os.Signal, 2)
	signal.Notify(s.signalChannel, syscall.SIGUSR1, syscall.SIGUSR2)
	go s.run()
	return s
}
