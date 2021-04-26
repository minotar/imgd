package main

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/op/go-logging"
)

type StashingWriter struct {
	buf bytes.Buffer
}

func (w *StashingWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func (w *StashingWriter) Unstash() string {
	return w.buf.String()
}

func testSetupSignals() *StashingWriter {
	sw := new(StashingWriter)
	logBackend := logging.NewLogBackend(sw, "", 0)
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	// Ensure we have debug log level
	logging.SetLevel(5, "")
	setupTestCache()
	return sw
}

func testDumpsProfile(t *testing.T, prefixText string, signal os.Signal) {
	sw := testSetupSignals()
	sh := new(SignalHandler)
	sh.handleSignal(signal)
	logOutput := sw.Unstash()
	t.Logf("Actual log output\n%s\nEnd actual log output", logOutput)
	if !strings.Contains(logOutput, prefixText) {
		t.Fatalf("Did not output '%s' in log", prefixText)
	}
	if strings.Count(logOutput, prefixText) > 1 {
		t.Fatalf("'%s' occurred multiple times in log", prefixText)
	}
	pprofLocation := logOutput[strings.Index(logOutput, prefixText)+len(prefixText):]
	pprofLocation = pprofLocation[:strings.Index(pprofLocation, "\n")]
	if _, err := os.Stat(pprofLocation); os.IsNotExist(err) {
		t.Fatal("Pprof output path " + pprofLocation + " does not exist!")
	} else {
		os.Remove(pprofLocation)
	}
}

func TestDumpsBlockProfile(t *testing.T) {
	testDumpsProfile(t, "Dumped block pprof to ", syscall.SIGUSR1)
}
func TestDumpsGoroutineProfile(t *testing.T) {
	testDumpsProfile(t, "Dumped goroutine pprof to ", syscall.SIGUSR2)
}
