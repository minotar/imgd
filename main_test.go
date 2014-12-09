package main

import (
	"github.com/op/go-logging"
	"testing"
)

type SilentWriter struct {
}

func (w SilentWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func TestNothing(t *testing.T) {

}

func BenchmarkSetup(b *testing.B) {
	logBackend := logging.NewLogBackend(SilentWriter{}, "", 0)
	setupConfig()
	setupLog(logBackend)
	setupCache()
}

func BenchmarkGetHead(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetHead(20)
	}
}

func BenchmarkGetHelm(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetHelm(20)
	}
}

func BenchmarkGetBody(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBody(20)
	}
}

func BenchmarkGetBust(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBust(20)
	}
}

func BenchmarkGetCube(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetCube(20)
	}
}
