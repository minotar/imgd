package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	//"time"

	//"github.com/minotar/minecraft"
	//"github.com/minotar/minecraft/mockminecraft"

	//"github.com/minotar/imgd/storage"
	"github.com/gorilla/mux"
	"github.com/minotar/minecraft/mockminecraft"
	. "github.com/smartystreets/goconvey/convey"
)

func setupTestServer() (string, func()) {

	r := Router{Mux: mux.NewRouter()}
	r.Bind()

	testServer := httptest.NewServer(imgdHandler(r.Mux))

	return testServer.URL, testServer.Close
}

func setupHTTPMockClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func addDash(uuid string) string {
	return uuid[0:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:32]
}

func TestHTTPMockVersion(t *testing.T) {
	Convey("Test web handlers with mock HTTP server", t, func() {
		testURL, shutdown := setupTestServer()
		defer shutdown()
		client := setupHTTPMockClient()

		Convey("Test homepage redirect", func() {
			resp, err := client.Get(testURL)
			So(err, ShouldBeNil)
			defer resp.Body.Close()
			So(resp.StatusCode, ShouldEqual, http.StatusFound)
			So(resp.Header.Get("Location"), ShouldEqual, config.Server.URL)
		})

		workingUsernames := []string{
			"clone1018",
			"citricsquid",
		}
		avatarHash := map[string]string{
			"clone1018":   "71bdaacd85812af7cd7f4226c6c91c43",
			"citricsquid": "91ae8bcc4f238ede2e234e685a7009fb",
		}
		cubeHash := map[string]string{
			"clone1018":   "a9055bca4a27c0ef7b2bc511e18c7576",
			"citricsquid": "0fe263aef4b749c58e42a9089a416154",
		}
		bustHash := map[string]string{
			"clone1018":   "725ddd51b55b15db173279694df43d5d",
			"citricsquid": "94d506fe4846b01d033c22ac509e1dc7",
		}
		bodyHash := map[string]string{
			"clone1018":   "3e8ee3e1ed24f685820625cf80c12432",
			"citricsquid": "eed549817c360c77ac2aa3201e51aa1c",
		}
		armorBustHash := map[string]string{
			"clone1018":   "725ddd51b55b15db173279694df43d5d",
			"citricsquid": "94d506fe4846b01d033c22ac509e1dc7",
		}
		armorBodyHash := map[string]string{
			"clone1018":   "3e8ee3e1ed24f685820625cf80c12432",
			"citricsquid": "eed549817c360c77ac2aa3201e51aa1c",
		}
		downloadHash := map[string]string{
			"clone1018":   "602026c3174c3ba63737e4b82675afa8",
			"citricsquid": "37afc5d1a03364dc554e2f5859ac461d",
		}
		resourceHashes := map[string]map[string]string{
			"avatar":     avatarHash,
			"helm":       avatarHash,
			"cube":       cubeHash,
			"bust":       bustHash,
			"body":       bodyHash,
			"armor/bust": armorBustHash,
			"armor/body": armorBodyHash,
			"skin":       downloadHash,
			"download":   downloadHash,
		}

		for _, tResource := range []string{"Avatar", "Helm", "Cube", "Bust", "Body", "Armor/Bust", "Armor/Body", "Skin", "Download"} {
			tResource = strings.ToLower(tResource)
			Convey("Test resource request for "+tResource, func() {
				for _, tUsername := range workingUsernames {
					uuid := mockminecraft.APIProfilesUUID[tUsername]
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[uuid])
					skinPath := skinURL.Path
					skinHash := mockminecraft.TexturesHash[skinPath]

					Convey("Username lookup for "+tUsername+" should return expected hash/Etag and this should then return StatusNotModified", func() {
						path := "/" + tResource + "/" + tUsername
						resp, err := client.Get(testURL + path)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, http.StatusOK)
						So(resp.Header.Get("Etag"), ShouldEqual, skinHash)

						respBytes, _ := ioutil.ReadAll(resp.Body)
						hasher := md5.New()
						hasher.Write(respBytes)
						respHash := hex.EncodeToString(hasher.Sum(nil))
						resp.Body.Close()
						So(respHash, ShouldEqual, resourceHashes[tResource][tUsername])

						req, _ := http.NewRequest("GET", testURL+path, nil)
						req.Header.Set("If-None-Match", skinHash)
						resp, err = client.Do(req)
						So(err, ShouldBeNil)
						resp.Body.Close()
						So(resp.StatusCode, ShouldEqual, http.StatusNotModified)
					})
				}

				for _, tUsername := range workingUsernames {
					uuid := mockminecraft.APIProfilesUUID[tUsername]
					dashUUID := addDash(uuid)
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[uuid])
					skinPath := skinURL.Path
					skinHash := mockminecraft.TexturesHash[skinPath]

					Convey("UUID lookup for "+dashUUID+" should redirect to"+uuid, func() {
						path := "/" + tResource + "/" + dashUUID
						resp, err := client.Get(testURL + path)
						So(err, ShouldBeNil)
						resp.Body.Close()
						So(resp.StatusCode, ShouldEqual, http.StatusMovedPermanently)
						So(resp.Header.Get("Location"), ShouldEqual, "/"+tResource+"/"+uuid)

						req, _ := http.NewRequest("GET", testURL+path, nil)
						req.Header.Set("If-None-Match", skinHash)
						resp, err = client.Do(req)
						So(err, ShouldBeNil)
						resp.Body.Close()
						So(resp.StatusCode, ShouldEqual, http.StatusMovedPermanently)
						So(resp.Header.Get("Location"), ShouldEqual, "/"+tResource+"/"+uuid)
					})
				}

				for _, tUsername := range workingUsernames {
					uuid := mockminecraft.APIProfilesUUID[tUsername]
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[uuid])
					skinPath := skinURL.Path
					skinHash := mockminecraft.TexturesHash[skinPath]

					Convey("Username lookup for "+uuid+" should return expected hash", func() {
						path := "/" + tResource + "/" + uuid
						resp, err := client.Get(testURL + path)
						So(err, ShouldBeNil)
						defer resp.Body.Close()
						So(resp.StatusCode, ShouldEqual, http.StatusOK)
						So(resp.Header.Get("Etag"), ShouldEqual, skinHash)

						req, _ := http.NewRequest("GET", testURL+path, nil)
						req.Header.Set("If-None-Match", skinHash)
						resp, err = client.Do(req)
						So(err, ShouldBeNil)
						resp.Body.Close()
						So(resp.StatusCode, ShouldEqual, http.StatusNotModified)
					})
				}
			})
		}
	})
}
