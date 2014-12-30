package main

import (
	"github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type SilentWriter struct {
}

func (w SilentWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func TestSetup(t *testing.T) {
	logBackend := logging.NewLogBackend(SilentWriter{}, "", 0)
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	setupCache()
}

func TestRenders(t *testing.T) {
	Convey("GetHead should return valid a image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetHead(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetHelm should return valid a image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetHelm(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetCube should return valid a image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetCube(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetBust should return valid a image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetBust(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetBody should return valida  image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetBody(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetArmorBust should return a valid image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetArmorBust(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetArmorBody should return a valid image", t, func() {
		skin := fetchSkin("clone1018")
		err := skin.GetArmorBody(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func BenchmarkSetup(b *testing.B) {
	logBackend := logging.NewLogBackend(SilentWriter{}, "", 0)
	stats = MakeStatsCollector()
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

func BenchmarkGetCube(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetCube(20)
	}
}

func BenchmarkGetBust(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBust(20)
	}
}

func BenchmarkGetBody(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBody(20)
	}
}

func BenchmarkGetArmorBust(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetArmorBust(20)
	}
}

func BenchmarkGetArmorBody(b *testing.B) {
	skin := fetchSkin("clone1018")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetArmorBody(20)
	}
}
