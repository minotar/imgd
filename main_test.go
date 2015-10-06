package main

import (
	"crypto/md5"
	"fmt"
	"github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
	"image"
	_ "image/png"
	"testing"
)

const (
	testUser = "clone1018"
)

type SilentWriter struct {
}

func (w SilentWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func hashRender(render image.Image) string {
	// md5 hash its pixels
	hasher := md5.New()
	hasher.Write(render.(*image.NRGBA).Pix)
	md5Hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// And return the Hash
	return md5Hash
}

func TestSetup(t *testing.T) {
	logBackend := logging.NewLogBackend(SilentWriter{}, "", 0)
	stats = MakeStatsCollector()
	setupConfig()
	setupLog(logBackend)
	setupCache()
}

func TestRenders(t *testing.T) {
	Convey("GetHead should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetHead(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "257972f3af185ce1f92200156d3bfe97")
	})

	Convey("GetHelm should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetHelm(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "257972f3af185ce1f92200156d3bfe97")
	})

	Convey("GetCube should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetCube(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "a253bb68f5ed938eb235ff1e3807940c")
	})

	Convey("GetBust should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetBust(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "1a4ba8567619350883923bde97626ae6")
	})

	Convey("GetBody should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetBody(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "987cc81d031386f429b12da2cd5c1ff2")
	})

	Convey("GetArmorBust should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetArmorBust(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "1a4ba8567619350883923bde97626ae6")
	})

	Convey("GetArmorBody should return a valid image", t, func() {
		skin := fetchSkin(testUser)
		err := skin.GetArmorBody(20)

		So(skin.Processed, ShouldNotBeNil)
		So(err, ShouldBeNil)

		hash := hashRender(skin.Processed)
		So(hash, ShouldEqual, "987cc81d031386f429b12da2cd5c1ff2")
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
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetHead(20)
	}
}

func BenchmarkGetHelm(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetHelm(20)
	}
}

func BenchmarkGetCube(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetCube(20)
	}
}

func BenchmarkGetBust(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBust(20)
	}
}

func BenchmarkGetBody(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetBody(20)
	}
}

func BenchmarkGetArmorBust(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetArmorBust(20)
	}
}

func BenchmarkGetArmorBody(b *testing.B) {
	skin := fetchSkin(testUser)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		skin.GetArmorBody(20)
	}
}
