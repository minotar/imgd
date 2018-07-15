package main

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/minotar/minecraft"
	"github.com/minotar/minecraft/mockminecraft"

	"github.com/minotar/imgd/storage"
	. "github.com/smartystreets/goconvey/convey"
)

/*
From mockminecraft (reference)

APIProfiles = map[string]string{
	"clone1018":        `{"id":"d9135e082f2244c89cb0bee234155292","name":"clone1018"}`,
	"lukegb":           `{"id":"2f3665cc5e29439bbd14cb6d3a6313a7","name":"lukegb"}`,
	"lukehandle":       `{"id":"5c115ca73efd41178213a0aff8ef11e0","name":"LukeHandle"}`,
	"citricsquid":      `{"id":"48a0a7e4d5594873a617dc189f76a8a1","name":"citricsquid"}`,
	"ratelimitapi":     `{"id":"00000000000000000000000000000000","name":"RateLimitAPI"}`,
	"ratelimitsession": `{"id":"00000000000000000000000000000001","name":"RateLimitSession"}`,
	"malformedapi":     `{"id":"00000000000000000000000000000002","name":"MalformedAPI`,
	"malformedsession": `{"id":"00000000000000000000000000000003","name":"MalformedSession"}`,
	"notexture":        `{"id":"00000000000000000000000000000004","name":"NoTexture"}`,
	"malformedtexprop": `{"id":"00000000000000000000000000000005","name":"MalformedTexProp"}`,
	"500api":           `{"id":"00000000000000000000000000000006","name":"500API"}`,
	"500session":       `{"id":"00000000000000000000000000000007","name":"500Session"}`,
	"malformedstex":    `{"id":"00000000000000000000000000000008","name":"MalformedSTex"}`,
	"malformedctex":    `{"id":"00000000000000000000000000000009","name":"MalformedCTex"}`,
	"404stexture":      `{"id":"00000000000000000000000000000010","name":"404STexture"}`,
	"404ctexture":      `{"id":"00000000000000000000000000000011","name":"404CTexture"}`,
	"rlsessionmojang":  `{"id":"00000000000000000000000000000012","name":"RLSessionMojang"}`,
	"rlsessions3":      `{"id":"00000000000000000000000000000013","name":"RLSessionS3"}`,
	"nousername":       `{"id":"00000000000000000000000000000014","notname":"NoUsername"}`,
	"204session":       `{"id":"00000000000000000000000000000015","name":"204Session"}`,
}
SessionProfiles = map[string]string{
	"d9135e082f2244c89cb0bee234155292": `{"id":"d9135e082f2244c89cb0bee234155292","name":"clone1018"[...]`,
	"2f3665cc5e29439bbd14cb6d3a6313a7": `{"id":"2f3665cc5e29439bbd14cb6d3a6313a7","name":"lukegb"[...]`,
	"5c115ca73efd41178213a0aff8ef11e0": `{"id":"5c115ca73efd41178213a0aff8ef11e0","name":"LukeHandle"[...]`,
	"48a0a7e4d5594873a617dc189f76a8a1": `{"id":"48a0a7e4d5594873a617dc189f76a8a1","name":"citricsquid"[...]`,
	"00000000000000000000000000000003": `{"id":"00000000000000000000000000000003","name":"MalformedSession"`,
	"00000000000000000000000000000004": `{"id":"00000000000000000000000000000004","name":"NoTexture"[...]`,
	"00000000000000000000000000000005": `{"id":"00000000000000000000000000000005","name":"MalformedTexProp"[...]`,
	"00000000000000000000000000000008": `{"id":"00000000000000000000000000000008","name":"MalformedSTex"[...]`,
	"00000000000000000000000000000009": `{"id":"00000000000000000000000000000009","name":"MalformedCTex"[...]`,
	"00000000000000000000000000000010": `{"id":"00000000000000000000000000000010","name":"404STexture"[...]`,
	"00000000000000000000000000000011": `{"id":"00000000000000000000000000000011","name":"404CTexture"[...]`,
	"00000000000000000000000000000014": `{"id":"00000000000000000000000000000014","properties":[]}`,
}
Textures = map[string]string{
	// clone1018 skin
	"cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649": `data`,
	// citricquid skin
	"e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f": `data`,
	// citricquid cape
	"c3af7fb821254664558f28361158ca73303c9a85e96e5251102958d7ed60c4a3": `data`,
	// Malformed
	"MalformedTexture": `data`,
	// Dud
	"404Texture": ``,
}
*/

// DelayedTransport is an http.RoundTripper that rewrites requests
// using the provided URL's Scheme and Host, and its Path as a prefix.
// Crucially, it adds a slight delay to help the singleflight test
// The Opaque field is untouched.
// If Transport is nil, http.DefaultTransport is used
type DelayedTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

// RoundTrip is used by the http.Client and rewrites the request to the testserver
func (t DelayedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// note that url.URL.ResolveReference doesn't work here
	// since t.u is an absolute url
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	time.Sleep(1 * time.Millisecond)
	//req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func TestUsernameLookup(t *testing.T) {

	Convey("test getUUID returns expected data from Username", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working Usernames should return positive results", func() {

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"clone1018", "lukegb", "lukehandle", "citricsquid"} {
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheUUID does not contain Username
					uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuid, ShouldBeBlank)

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetAPIProfile"]

					// Perform the API Request
					uuid, err = getUUID(tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, mockminecraft.APIProfilesUUID[tUsername])

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origRequests+1)

					// Verify cacheUUID is now populated
					uuid, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, mockminecraft.APIProfilesUUID[tUsername])

					// Verify the cache is used and the APIRequested has not increased
					uuid, err = getUUID(tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, mockminecraft.APIProfilesUUID[tUsername])

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origRequests+1)

					// Verify if we blank the record in cacheUUID, a lookup will occur
					// Blanking is an explicit behaviour which may be used when we
					// want to force a "miss"
					err = storage.InsertKV(cache["cacheUUID"], tUsername, "", uuidErrorTTL)
					So(err, ShouldBeNil)

					// Verify we performed an API Request
					uuid, err = getUUID(tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, mockminecraft.APIProfilesUUID[tUsername])

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origRequests+2)

				})
			}
		})

		Convey("Broken Usernames should return negative results", func() {

			brokenUsernameResp := map[string]string{
				"ratelimitapi": metaRateLimitCode,
				"204api":       metaUnknownCode,
				"malformedapi": metaErrorCode,
				"500api":       metaErrorCode,
			}
			brokenUsernameError := map[string]string{
				"ratelimitapi": "rate limited",
				"204api":       "user not found",
				"malformedapi": "UUID lookup failed",
				"500api":       "UUID lookup failed",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for tUsername := range brokenUsernameResp {
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheUUID does not contain Username
					uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuid, ShouldBeBlank)

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetAPIProfile"]

					uuid, err = getUUID(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUsernameError[tUsername])
					So(uuid, ShouldEqual, "")

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origRequests+1)

					// Verify cacheUUID is now populated
					uuid, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, brokenUsernameResp[tUsername])

					// Verify the cache is used and the APIRequested has not increased
					uuid, err = getUUID(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUsernameError[tUsername])
					So(uuid, ShouldEqual, "")

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origRequests+1)
				})
			}
		})
	})
}

func TestSessionProfileLookup(t *testing.T) {

	Convey("test pullSessionProfile returns expected data from UUID", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working UUIDs should return positive results", func() {

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"clone1018", "lukegb", "lukehandle", "citricsquid"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					user := &mcUser{}
					user.UUID = tUUID

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetSessionProfile"]

					err := user.pullSessionProfile()
					So(err, ShouldBeNil)
					So(user.UUID, ShouldEqual, tUUID)
					So(user.Username, ShouldEqual, tUsername)
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
					So(user.Textures.SkinPath, ShouldEqual, skinURL.Path)

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)
				})
			}
		})

		Convey("Broken Usernames should return negative results", func() {

			brokenUUIDError := map[string]string{
				"00000000000000000000000000000001": "unable to GetSessionProfile: rate limited",
				"00000000000000000000000000000003": "decoding GetSessionProfile failed: unexpected EOF",
				"00000000000000000000000000000004": "unable to DecodeTextureProperty: no textures property",
				"00000000000000000000000000000005": "unable to DecodeTextureProperty: unexpected EOF",
				"00000000000000000000000000000007": "unable to GetSessionProfile: apiRequest HTTP 500 Internal Server Error",
				"00000000000000000000000000000014": "pullSessionProfile unable to grab Username",
				"00000000000000000000000000000015": "unable to GetSessionProfile: user not found",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"ratelimitsession", "malformedsession", "notexture", "malformedtexprop", "500session", "nousername", "204session"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					user := &mcUser{}
					user.UUID = tUUID

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetSessionProfile"]

					err := user.pullSessionProfile()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUUIDError[tUUID])
					So(user.UUID, ShouldEqual, tUUID)
					So(user.Textures.SkinPath, ShouldBeEmpty)

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)
				})
			}
		})
	})
}

func TestUUIDLookup(t *testing.T) {

	Convey("test getUser returns expected data from UUID", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working UUIDs should return positive results", func() {

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UserData
			for _, tUsername := range []string{"clone1018", "lukegb", "lukehandle", "citricsquid"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					user := &mcUser{}
					// Verify cacheUserData does not contain User
					err := storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user, ShouldResemble, &mcUser{})

					// Verify cacheUUID does not contain UUID
					uuidRet, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuidRet, ShouldBeEmpty)

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetSessionProfile"]

					user = &mcUser{}
					user, err = getUser(tUUID)
					So(err, ShouldBeNil)
					So(user.UUID, ShouldEqual, tUUID)
					So(user.Username, ShouldEqual, tUsername)
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
					So(user.Textures.SkinPath, ShouldEqual, skinURL.Path)
					So(user.Timestamp.IsZero(), ShouldBeFalse)
					// We will compare the other returned data against the first (which should be cached)
					userRet := *user
					// We need to remove the "Monotonic Clocks" values which are not added to the cache
					userRet.Timestamp = userRet.Timestamp.Round(0)

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)

					// Verify cacheUserData is now populated
					user = &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldBeNil)
					So(user, ShouldResemble, &userRet)

					// Verify cacheUUID is now populated
					time.Sleep(time.Duration(1) * time.Millisecond)
					uuidRet, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuidRet, ShouldEqual, tUUID)

					// Verify the cache is used and the APIRequested has not increased
					user = &mcUser{}
					user, err = getUser(tUUID)
					So(err, ShouldBeNil)
					So(user, ShouldResemble, &userRet)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)
				})
			}
		})

		Convey("Broken UUIDs should return negative results", func() {

			brokenUUIDResp := map[string]string{
				"00000000000000000000000000000001": metaRateLimitCode,
				"00000000000000000000000000000003": metaErrorCode,
				"00000000000000000000000000000004": metaErrorCode,
				"00000000000000000000000000000005": metaErrorCode,
				"00000000000000000000000000000007": metaErrorCode,
				"00000000000000000000000000000014": metaErrorCode,
				"00000000000000000000000000000015": metaUnknownCode,
			}
			brokenUUIDError := map[string]string{
				"00000000000000000000000000000001": "rate limited",
				"00000000000000000000000000000003": "SessionProfile lookup failed",
				"00000000000000000000000000000004": "SessionProfile lookup failed",
				"00000000000000000000000000000005": "SessionProfile lookup failed",
				"00000000000000000000000000000007": "SessionProfile lookup failed",
				"00000000000000000000000000000014": "SessionProfile lookup failed",
				"00000000000000000000000000000015": "user not found",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the UserData
			for _, tUsername := range []string{"ratelimitsession", "malformedsession", "notexture", "malformedtexprop", "500session", "nousername", "204session"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					user := &mcUser{}
					// Verify cacheUserData does not contain User
					err := storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user, ShouldResemble, &mcUser{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["GetSessionProfile"]

					user = &mcUser{}
					user, err = getUser(tUUID)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUUIDError[tUUID])
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{
						User: minecraft.User{UUID: tUUID},
					})

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)

					// Verify the cacheUserData is now populated
					user = &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldBeNil)
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{
						User:     minecraft.User{UUID: tUUID},
						Textures: textures{SkinPath: brokenUUIDResp[tUUID]},
					})

					// Verify the cache is used and the APIRequested has not increased
					user = &mcUser{}
					user, err = getUser(tUUID)
					So(err, ShouldNotBeNil)
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(err.Error(), ShouldEqual, brokenUUIDError[tUUID])
					So(user, ShouldResemble, &mcUser{
						User: minecraft.User{UUID: tUUID},
					})

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)

					// Verify cacheUUID is not populated
					uuidRet, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuidRet, ShouldBeEmpty)
				})
			}
		})
	})
}

func TestSkinLookup(t *testing.T) {

	Convey("test getSkin returns expected data from SkinPath", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working SkinPaths should return positive results", func() {

			// Loop though the 2 Usernames with Skins in mockminecraft and compare against the correct Skin
			for _, tUsername := range []string{"clone1018", "citricsquid"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
				skinPath := skinURL.Path
				tHash := mockminecraft.TexturesHash[skinPath]
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					// Verify cacheSkin does not contain Skin
					skin, err := RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(skin, ShouldResemble, minecraft.Skin{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["TextureFetch"]

					skin, err = getSkin(skinPath)
					So(err, ShouldBeNil)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origRequests+1)

					// Verify cacheSkin is now populated
					skin, err = RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldBeNil)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					// Verify the cache is used and the APIRequested has not increased
					skin, err = getSkin(skinPath)
					So(err, ShouldBeNil)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origRequests+1)
				})
			}
		})

		Convey("Broken SkinPaths should return negative results", func() {

			brokenUUIDError := map[string]string{
				"00000000000000000000000000000008": "unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data",
				"00000000000000000000000000000010": "unable to Fetch Texture: apiRequest HTTP 404 Not Found",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the UserData
			for _, tUsername := range []string{"malformedstex", "404stexture"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
				skinPath := skinURL.Path
				Convey("Test lookup for "+tUUID+" ("+tUsername+")", func() {
					// Verify cacheUserData does not contain User
					skin, err := RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(skin, ShouldResemble, minecraft.Skin{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origRequests := stats.info.APIRequested["TextureFetch"]

					skin, err = getSkin(skinPath)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "getSkin failed: "+brokenUUIDError[tUUID])
					So(skin.Hash, ShouldBeEmpty)
					So(skin.Image, ShouldBeNil)

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origRequests+1)

					// Verify cacheSkinTransient is now populated
					state, err := storage.RetrieveKV(cache["cacheSkinTransient"], skinPath)
					So(err, ShouldBeNil)
					So(state, ShouldEqual, brokenUUIDError[tUUID])

					// Verify the cache is used and the APIRequested has not increased
					skin, err = getSkin(skinPath)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "getSkin previously failed: "+brokenUUIDError[tUUID])
					So(skin.Hash, ShouldBeEmpty)
					So(skin.Image, ShouldBeNil)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origRequests+1)
				})
			}
		})

	})
}

func TestSingleFlight(t *testing.T) {

	Convey("test singleflight blocks duplicate requests to the API", t, func() {
		// reset Caches
		setupTestCache()
		origClient := mcClient.Client

		u, err := url.Parse(mockminecraft.TestURL)
		if err != nil {
			t.Fail()
		}

		rt := DelayedTransport{URL: u}
		mcClient.Client = &http.Client{Transport: rt}

		Convey("Concurrent wrapUUIDLookup calls should perform a single request", func() {
			tUsername := "lukehandle"
			tUUID := mockminecraft.APIProfilesUUID[tUsername]

			// Get the current APIRequested count and CacheUserData Hits to compare against
			time.Sleep(time.Duration(1) * time.Millisecond)
			origRequests := stats.info.APIRequested["GetSessionProfile"]
			origCacheHits := stats.info.CacheUserData.Hits

			go func() {
				wrapUUIDLookup(tUUID)
			}()
			go func() {
				wrapUUIDLookup(tUUID)
			}()

			user, err := wrapUUIDLookup(tUUID)

			So(err, ShouldBeNil)
			So(user.Username, ShouldEqual, tUsername)
			So(user.UUID, ShouldEqual, tUUID)

			// Verify we made an APIRequest
			time.Sleep(time.Duration(1) * time.Millisecond)
			So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)
			So(stats.info.CacheUserData.Hits, ShouldEqual, origCacheHits)
		})

		Convey("Concurrent wrapUsernameLookup calls should perform a single request", func() {
			tUsername := "lukehandle"
			tUUID := mockminecraft.APIProfilesUUID[tUsername]

			// Get the current APIRequested count and CacheUserData Hits to compare against
			time.Sleep(time.Duration(1) * time.Millisecond)
			origRequests := stats.info.APIRequested["GetSessionProfile"]
			origCacheHits := stats.info.CacheUserData.Hits

			go func() {
				wrapUsernameLookup(tUsername)
			}()
			go func() {
				wrapUsernameLookup(tUsername)
			}()

			user, err := wrapUsernameLookup(tUsername)

			So(err, ShouldBeNil)
			So(user.Username, ShouldEqual, tUsername)
			So(user.UUID, ShouldEqual, tUUID)

			// Verify we made an APIRequest
			time.Sleep(time.Duration(1) * time.Millisecond)
			So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origRequests+1)
			So(stats.info.CacheUserData.Hits, ShouldEqual, origCacheHits)
		})
		mcClient.Client = origClient
	})
}

func TestWrapUsernameLookup(t *testing.T) {

	Convey("test wrapUsernameLookup returns expected data from Username", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working Usernames should return positive results", func() {

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"clone1018", "lukegb", "lukehandle", "citricsquid"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheUUID does not contain Username
					uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuid, ShouldBeBlank)

					// Verify cacheUserData does not contain User
					user := &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user, ShouldResemble, &mcUser{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
					origSessionRequests := stats.info.APIRequested["GetSessionProfile"]

					// Perform the API Request
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldBeNil)
					So(user.UUID, ShouldEqual, tUUID)
					So(user.Username, ShouldEqual, tUsername)
					skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
					So(user.Textures.SkinPath, ShouldEqual, skinURL.Path)
					So(user.Timestamp.IsZero(), ShouldBeFalse)
					// We will compare the other returned data against the first (which should be cached)
					userRet := *user
					// We need to remove the "Monotonic Clocks" values which are not added to the cache
					userRet.Timestamp = userRet.Timestamp.Round(0)

					// Verify we made APIRequests
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)

					// Verify cacheUUID is now populated
					uuid, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, tUUID)

					// Verify cacheUserData is now populated
					user = &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldBeNil)
					So(user, ShouldResemble, &userRet)

					// Verify the cache is used and the APIRequested has not increased
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldBeNil)
					So(user, ShouldResemble, &userRet)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)
				})
			}
		})

		Convey("Broken Usernames should return negative results", func() {

			brokenUsernameResp := map[string]string{
				"ratelimitapi": metaRateLimitCode,
				"204api":       metaUnknownCode,
				"malformedapi": metaErrorCode,
				"500api":       metaErrorCode,
			}
			brokenUsernameError := map[string]string{
				"ratelimitapi": "rate limited",
				"204api":       "user not found",
				"malformedapi": "UUID lookup failed",
				"500api":       "UUID lookup failed",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for tUsername := range brokenUsernameResp {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheUUID does not contain Username
					uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuid, ShouldBeBlank)

					// Verify cacheUserData does not contain User
					user := &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user, ShouldResemble, &mcUser{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
					origSessionRequests := stats.info.APIRequested["GetSessionProfile"]

					// Perform the API Request
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUsernameError[tUsername])
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{})

					// Verify we made an APIRequest
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests)

					// Verify cacheUUID is now populated
					uuid, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, brokenUsernameResp[tUsername])

					// Verify the cacheUserData is now populated
					user = &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{})

					// Verify the cache is used and the APIRequested has not increased
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUsernameError[tUsername])
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{})

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests)
				})
			}
		})

		Convey("Broken UUIDs should return negative results", func() {

			brokenUUIDResp := map[string]string{
				"00000000000000000000000000000001": metaRateLimitCode,
				"00000000000000000000000000000003": metaErrorCode,
				"00000000000000000000000000000004": metaErrorCode,
				"00000000000000000000000000000005": metaErrorCode,
				"00000000000000000000000000000007": metaErrorCode,
				"00000000000000000000000000000014": metaErrorCode,
				"00000000000000000000000000000015": metaUnknownCode,
			}
			brokenUUIDError := map[string]string{
				"00000000000000000000000000000001": "rate limited",
				"00000000000000000000000000000003": "SessionProfile lookup failed",
				"00000000000000000000000000000004": "SessionProfile lookup failed",
				"00000000000000000000000000000005": "SessionProfile lookup failed",
				"00000000000000000000000000000007": "SessionProfile lookup failed",
				"00000000000000000000000000000014": "SessionProfile lookup failed",
				"00000000000000000000000000000015": "user not found",
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"ratelimitsession", "malformedsession", "notexture", "malformedtexprop", "500session", "nousername", "204session"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheUUID does not contain Username
					uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(uuid, ShouldBeBlank)

					// Verify cacheUserData does not contain User
					user := &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(user, ShouldResemble, &mcUser{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
					origSessionRequests := stats.info.APIRequested["GetSessionProfile"]

					// Perform the API Request
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUUIDError[tUUID])
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{
						User: minecraft.User{UUID: tUUID},
					})
					// Verify we made APIRequests
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)

					// Verify cacheUUID is now populated
					uuid, err = storage.RetrieveKV(cache["cacheUUID"], tUsername)
					So(err, ShouldBeNil)
					So(uuid, ShouldEqual, tUUID)

					// Verify the cacheUserData is now populated
					user = &mcUser{}
					err = storage.RetrieveGob(cache["cacheUserData"], tUUID, user)
					So(err, ShouldBeNil)
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{
						User:     minecraft.User{UUID: tUUID},
						Textures: textures{SkinPath: brokenUUIDResp[tUUID]},
					})

					// Verify the cache is used and the APIRequested has not increased
					user, err = wrapUsernameLookup(tUsername)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, brokenUUIDError[tUUID])
					So(user.Timestamp.IsZero(), ShouldBeTrue)
					So(user, ShouldResemble, &mcUser{
						User: minecraft.User{UUID: tUUID},
					})

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)
				})
			}
		})

		Convey("Wrongly cached Username should return positive results", func() {
			tUsername := "lukehandle"
			tUUID := mockminecraft.APIProfilesUUID[tUsername]
			bUUID := mockminecraft.APIProfilesUUID["lukegb"]

			// Populate the cacheUUID with the incorrect UUID
			err := storage.InsertKV(cache["cacheUUID"], tUsername, bUUID, uuidTTL)
			So(err, ShouldBeNil)

			// Get the current APIRequested count to compare against
			time.Sleep(time.Duration(1) * time.Millisecond)
			origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
			origSessionRequests := stats.info.APIRequested["GetSessionProfile"]
			origStaleErrors := stats.info.Errored["APIProfileStale"]

			// Perform the API Request
			user, err := wrapUsernameLookup(tUsername)
			So(err, ShouldBeNil)
			So(user.UUID, ShouldNotEqual, bUUID)
			So(user.UUID, ShouldEqual, tUUID)
			So(user.Username, ShouldEqual, tUsername)
			skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
			So(user.Textures.SkinPath, ShouldEqual, skinURL.Path)
			So(user.Timestamp.IsZero(), ShouldBeFalse)
			// We will compare the other returned data against the first (which should be cached)
			userRet := *user
			// We need to remove the "Monotonic Clocks" values which are not added to the cache
			userRet.Timestamp = userRet.Timestamp.Round(0)

			// Verify we made APIRequests
			time.Sleep(time.Duration(1) * time.Millisecond)
			So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
			So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+2)
			So(stats.info.Errored["APIProfileStale"], ShouldEqual, origStaleErrors+1)

			// Verify cacheUUID is now populated correctly
			uuid, err := storage.RetrieveKV(cache["cacheUUID"], tUsername)
			So(err, ShouldBeNil)
			So(uuid, ShouldEqual, tUUID)
		})
	})
}

func TestFetchUsernameSkin(t *testing.T) {

	Convey("test fetchUsernameSkin returns expected data from Username", t, func() {
		// reset Caches
		setupTestCache()

		Convey("Working Usernames should return positive results", func() {

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for _, tUsername := range []string{"clone1018", "citricsquid"} {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
				skinPath := skinURL.Path
				tHash := mockminecraft.TexturesHash[skinPath]
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheSkin does not contain Skin
					mSkin, err := RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(mSkin, ShouldResemble, minecraft.Skin{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
					origSessionRequests := stats.info.APIRequested["GetSessionProfile"]
					origTextureRequests := stats.info.APIRequested["TextureFetch"]

					// Perform the API Request
					skin := fetchUsernameSkin(tUsername)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					// Verify we made APIRequests
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origTextureRequests+1)

					// Verify cacheSkin is now populated
					mSkin, err = RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldBeNil)
					So(mSkin.Hash, ShouldNotBeEmpty)
					So(mSkin.Hash, ShouldEqual, tHash)

					// Verify the cache is used and the APIRequested has not increased
					skin = fetchUsernameSkin(tUsername)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+1)
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origTextureRequests+1)
				})
			}
		})

		Convey("Broken Usernames should return negative results", func() {

			brokenUsernameProfileReq := map[string]uint{
				"ratelimitapi":     0,
				"204api":           0,
				"malformedapi":     0,
				"500api":           0,
				"ratelimitsession": 1,
				"malformedsession": 1,
				"notexture":        1,
				"malformedtexprop": 1,
				"500session":       1,
				"nousername":       1,
				"204session":       1,
			}

			// Loop though a selection of the Usernames from mockminecraft and compare against the correct UUID
			for tUsername := range brokenUsernameProfileReq {
				tUUID := mockminecraft.APIProfilesUUID[tUsername]
				skinURL, _ := url.Parse(mockminecraft.SessionProfilesSkinPath[tUUID])
				skinPath := skinURL.Path
				tHash := minecraft.SteveHash
				Convey("Test lookup for "+tUsername+" and caching", func() {
					// Verify cacheSkin does not contain Skin
					mSkin, err := RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(mSkin, ShouldResemble, minecraft.Skin{})

					// Get the current APIRequested count to compare against
					time.Sleep(time.Duration(1) * time.Millisecond)
					origAPIRequests := stats.info.APIRequested["GetAPIProfile"]
					origSessionRequests := stats.info.APIRequested["GetSessionProfile"]
					origTextureRequests := stats.info.APIRequested["TextureFetch"]

					// Perform the API Request
					skin := fetchUsernameSkin(tUsername)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					// Verify we made APIRequests
					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+brokenUsernameProfileReq[tUsername])
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origTextureRequests)

					// Verify cacheSkin does not contain Skin
					mSkin, err = RetrieveSkin(cache["cacheSkin"], skinPath)
					So(err, ShouldEqual, storage.ErrNotFound)
					So(mSkin, ShouldResemble, minecraft.Skin{})

					// Verify the cache is used and the APIRequested has not increased
					skin = fetchUsernameSkin(tUsername)
					So(skin.Hash, ShouldNotBeEmpty)
					So(skin.Hash, ShouldEqual, tHash)

					time.Sleep(time.Duration(1) * time.Millisecond)
					So(stats.info.APIRequested["GetAPIProfile"], ShouldEqual, origAPIRequests+1)
					So(stats.info.APIRequested["GetSessionProfile"], ShouldEqual, origSessionRequests+brokenUsernameProfileReq[tUsername])
					So(stats.info.APIRequested["TextureFetch"], ShouldEqual, origTextureRequests)
				})
			}
		})

	})
}

// Too many lines
