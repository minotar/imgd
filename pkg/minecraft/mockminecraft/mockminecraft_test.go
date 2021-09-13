package mockminecraft

import (
	"errors"
	"net/http"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var client *http.Client

func doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("unable to create request: " + err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("unable to GET URL: " + err.Error())
	}
	return resp, nil
}

func TestMain(m *testing.M) {
	rt, shutdown := Setup(ReturnMux())
	client = &http.Client{Transport: rt}
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestServer(t *testing.T) {

	Convey("Test bad GET request", t, func() {

		Convey("Bad Request", func() {
			_, err := doRequest("::")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to create request: parse ::: missing protocol scheme")
		})

	})

	Convey("Test Profile Response Codes", t, func() {

		apiProfiles := map[string]int{
			"clone1018":      http.StatusOK,
			"lukegb":         http.StatusOK,
			"LukeHandle":     http.StatusOK,
			"citricsquid":    http.StatusOK,
			"skmkj88200aklk": http.StatusNoContent,
			"RateLimitAPI":   http.StatusTooManyRequests,
			"500API":         http.StatusInternalServerError,
		}

		for name, respCode := range apiProfiles {
			Convey("Profile "+name+" responds with expected response code", func() {
				resp, err := doRequest("http://example.com/users/profiles/minecraft/" + name)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, respCode)
			})
		}
	})

	Convey("Test SessionProfile Response Codes", t, func() {

		sessionProfiles := map[string]int{
			"d9135e082f2244c89cb0bee234155292": http.StatusOK,
			"2f3665cc5e29439bbd14cb6d3a6313a7": http.StatusOK,
			"5c115ca73efd41178213a0aff8ef11e0": http.StatusOK,
			"48a0a7e4d5594873a617dc189f76a8a1": http.StatusOK,
			"00000000000000000000000000000001": http.StatusTooManyRequests,
			"00000000000000000000000000000012": http.StatusTooManyRequests,
			"00000000000000000000000000000013": http.StatusTooManyRequests,
			"00000000000000000000000000000007": http.StatusInternalServerError,
		}

		for uuid, respCode := range sessionProfiles {
			Convey("SessionProfile "+uuid+" responds with expected response code", func() {
				resp, err := doRequest("http://example.com/session/minecraft/profile/" + uuid)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, respCode)
			})
		}
	})

	Convey("Test Texture Response Codes", t, func() {

		textures := map[string]int{
			"cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649":  http.StatusOK,
			"e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f":  http.StatusOK,
			"c3af7fb821254664558f28361158ca73303c9a85e96e5251102958d7ed60c4a3": http.StatusOK,
			"404Texture": http.StatusNotFound,
		}

		for skinPath, respCode := range textures {
			Convey("SkinPath "+skinPath+" responds with expected response code", func() {
				resp, err := doRequest("http://example.com/texture/" + skinPath)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, respCode)
			})
		}
	})

}

func TestVariableChanging(t *testing.T) {

	urls := []string{
		"http://example.com/users/profiles/minecraft/mockminecraft_test",
		"http://example.com/session/minecraft/profile/10000000000000000000000000000000",
		"http://example.com/texture/mockminecraft_test",
	}

	Convey("Test 404 before adding to Maps", t, func() {

		for _, url := range urls {
			Convey("URL "+url+" responds with 204 (not found)", func() {
				resp, err := doRequest(url)

				So(err, ShouldBeNil)
				if url != "http://example.com/texture/mockminecraft_test" {
					So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				} else {
					So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
				}
			})
		}
	})

	APIProfiles["mockminecraft_test"] = `{"id":"10000000000000000000000000000000","name":"mockminecraft_test"}`
	SessionProfiles["10000000000000000000000000000000"] = `eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjEwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwIiwicHJvZmlsZU5hbWUiOiJtb2NrbWluZWNyYWZ0X3Rlc3QiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvbW9ja21pbmVjcmFmdF90ZXN0In19fQ==`
	Textures["/texture/mockminecraft_test"] = `iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVR4nGP6DwABBQECz6AuzQAAAABJRU5ErkJggg==`

	Convey("Test 200 after adding to Maps", t, func() {

		for _, url := range urls {
			Convey("URL "+url+" responds with 200", func() {
				resp, err := doRequest(url)

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		}
	})

	CreateMaps()

	Convey("Test 404 after resetting Maps", t, func() {

		for _, url := range urls {
			Convey("URL "+url+" responds with 204 (not found)", func() {
				resp, err := doRequest(url)

				So(err, ShouldBeNil)
				if url != "http://example.com/texture/mockminecraft_test" {
					So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				} else {
					So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
				}
			})
		}
	})

}
