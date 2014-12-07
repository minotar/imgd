package main

import (
	"testing"
)

func TestNothing(t *testing.T) {

}

func BenchmarkGetHead(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		GetHead(skin)
	}
}

func BenchmarkGetHelm(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		GetHelm(skin)
	}
}

func BenchmarkGetBody(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		GetBody(skin)
	}
}

func BenchmarkGetBust(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		GetBust(skin)
	}
}
