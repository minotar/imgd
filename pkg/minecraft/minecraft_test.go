// minecraft_test.go
package minecraft

import (
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/minotar/imgd/pkg/minecraft/mockminecraft"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mcTest *Minecraft
	mcProd *Minecraft
)

func TestMain(m *testing.M) {
	mux := mockminecraft.ReturnMux()
	rt, shutdown := mockminecraft.Setup(mux)

	cfg := &Config{
		UsernameAPIConfig: UsernameAPIConfig{
			SkinURL: "http://skins.example.net/skins/",
			CapeURL: "http://skins.example.net/capes/",
		},
	}
	mcTest = NewMinecraft(cfg)
	mcTest.Client = &http.Client{Transport: rt}
	mcProd = NewDefaultMinecraft()

	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestRegexs(t *testing.T) {

	Convey("Regexs compile", t, func() {
		var err error

		_, err = regexp.Compile("^" + ValidUsernameRegex + "$")
		So(err, ShouldBeNil)

		_, err = regexp.Compile("^" + ValidUUIDRegex + "$")
		So(err, ShouldBeNil)

		_, err = regexp.Compile("^" + ValidUsernameOrUUIDRegex + "$")
		So(err, ShouldBeNil)
	})

	Convey("Regexs work", t, func() {
		invalidUsernames := []string{"d9135e082f2244c89cb0bee234155292", "_-proscope-_", "PeriScopeButTooLong"}
		validUsernames := []string{"clone1018", "lukegb", "Wooxye"}

		invalidUUIDs := []string{"clone1018", "d9135e082f2244c8-9cb0-bee234155292"}
		validUUIDs := []string{"d9135e082f2244c89cb0bee234155292", "d9135e08-2f22-44c8-9cb0-bee234155292"}

		validUsernamesOrUUIDs := append(validUsernames, validUUIDs...)
		possiblyInvalidUsernamesOrUUIDs := append(invalidUsernames, invalidUUIDs...)

		Convey("Username regex works", func() {
			for _, validUsername := range validUsernames {
				So(RegexUsername.MatchString(validUsername), ShouldBeTrue)
			}

			for _, invalidUsername := range invalidUsernames {
				So(RegexUsername.MatchString(invalidUsername), ShouldBeFalse)
			}
		})

		Convey("UUID regex works", func() {
			for _, validUUID := range validUUIDs {
				So(RegexUUID.MatchString(validUUID), ShouldBeTrue)
			}

			for _, invalidUUID := range invalidUUIDs {
				So(RegexUUID.MatchString(invalidUUID), ShouldBeFalse)
			}
		})

		Convey("Username-or-UUID regex works", func() {
			for _, validThing := range validUsernamesOrUUIDs {
				So(RegexUsernameOrUUID.MatchString(validThing), ShouldBeTrue)
			}

			for _, possiblyInvalidThing := range possiblyInvalidUsernamesOrUUIDs {
				resultOne := RegexUsername.MatchString(possiblyInvalidThing)
				resultTwo := RegexUUID.MatchString(possiblyInvalidThing)
				expectedResult := resultOne || resultTwo

				So(RegexUsernameOrUUID.MatchString(possiblyInvalidThing), ShouldEqual, expectedResult)
			}
		})

	})

}

func TestExtra(t *testing.T) {

	Convey("Test bad GET requests", t, func() {

		Convey("apiRequest Bad Request", func() {
			_, err := mcProd.apiRequest("::")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to create request: parse \"::\": missing protocol scheme")
		})

		Convey("apiRequest Bad GET", func() {
			_, err := mcProd.apiRequest("//dummy_url")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to GET URL: Get \"//dummy_url\": unsupported protocol scheme \"\"")
		})

		Convey("t.Fetch Bad GET", func() {
			texture := &Texture{URL: "//dummy_url", Mc: mcProd}

			err := texture.Fetch()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to Fetch Texture: unable to GET URL: Get \"//dummy_url\": unsupported protocol scheme \"\"")
		})

	})

}
