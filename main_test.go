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
		skin.GetHead()
	}
}

func BenchmarkGetHelm(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetHelm()
	}
}

func BenchmarkGetBody(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBody()
	}
}

func BenchmarkGetBust(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		GetBust(skin)
	}
}
