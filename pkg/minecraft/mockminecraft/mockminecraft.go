package mockminecraft

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

var (
	// APIProfiles is a map of Username -> JSON return with the UUID
	APIProfiles map[string]string
	// APIProfilesUUID is a map of Username -> UUID (for performing through a range in a test)
	APIProfilesUUID map[string]string
	// SessionProfiles is a maps of UUID -> JSON return with the Texture Paths
	SessionProfiles map[string]string
	// SessionProfilesSkinPath is a maps of UUID -> SkinPaths (for performing through a range in a test)
	SessionProfilesSkinPath map[string]string
	// Textures are a map of SkinPaths -> Base64 encoded Texture
	Textures map[string]string
	// Textures are a map of SkinPaths -> Hash of Texture
	TexturesHash map[string]string
	// TestURL points to the testserver
	TestURL string
)

// CreateMaps adds test data to the Maps and can be used to reset them
func CreateMaps() {

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

	APIProfilesUUID = map[string]string{
		"clone1018":        "d9135e082f2244c89cb0bee234155292",
		"lukegb":           "2f3665cc5e29439bbd14cb6d3a6313a7",
		"lukehandle":       "5c115ca73efd41178213a0aff8ef11e0",
		"citricsquid":      "48a0a7e4d5594873a617dc189f76a8a1",
		"ratelimitapi":     "00000000000000000000000000000000",
		"ratelimitsession": "00000000000000000000000000000001",
		"malformedapi":     "00000000000000000000000000000002",
		"malformedsession": "00000000000000000000000000000003",
		"notexture":        "00000000000000000000000000000004",
		"malformedtexprop": "00000000000000000000000000000005",
		"500api":           "00000000000000000000000000000006",
		"500session":       "00000000000000000000000000000007",
		"malformedstex":    "00000000000000000000000000000008",
		"malformedctex":    "00000000000000000000000000000009",
		"404stexture":      "00000000000000000000000000000010",
		"404ctexture":      "00000000000000000000000000000011",
		"rlsessionmojang":  "00000000000000000000000000000012",
		"rlsessions3":      "00000000000000000000000000000013",
		"nousername":       "00000000000000000000000000000014",
		"204session":       "00000000000000000000000000000015",
	}

	SessionProfiles = map[string]string{
		"d9135e082f2244c89cb0bee234155292": `{"id":"d9135e082f2244c89cb0bee234155292","name":"clone1018","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6ImQ5MTM1ZTA4MmYyMjQ0Yzg5Y2IwYmVlMjM0MTU1MjkyIiwicHJvZmlsZU5hbWUiOiJjbG9uZTEwMTgiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvY2Q5Y2E1NWU5ODYyZjAwM2ViZmExODcyYTkyNDRhZDVmNzIxZDZiOWU2ODgzZGQxZDQyZjg3ZGFlMTI3NjQ5In19fQ=="}]}`,
		"2f3665cc5e29439bbd14cb6d3a6313a7": `{"id":"2f3665cc5e29439bbd14cb6d3a6313a7","name":"lukegb","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjJmMzY2NWNjNWUyOTQzOWJiZDE0Y2I2ZDNhNjMxM2E3IiwicHJvZmlsZU5hbWUiOiJsdWtlZ2IiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvYjU4YTMxOGEzYWI3YTA3NzYzNzhhMjhiYjI5ZTQyODdhODU0NDhhYmMzOTgxYTc5ZjQwMWUyYjdkZGYyMyJ9fX0="}]}`,
		"5c115ca73efd41178213a0aff8ef11e0": `{"id":"5c115ca73efd41178213a0aff8ef11e0","name":"LukeHandle","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjVjMTE1Y2E3M2VmZDQxMTc4MjEzYTBhZmY4ZWYxMWUwIiwicHJvZmlsZU5hbWUiOiJMdWtlSGFuZGxlIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlLzZmNzM2YjRjM2UyMjg2Y2ZhZDliMGQ3MzhmZDdkOTYzMGQ5ZTBhMjc3MjFiNzU4NmU0MjNjZWJjZTQyMGRhIn19fQ=="}]}`,
		"48a0a7e4d5594873a617dc189f76a8a1": `{"id":"48a0a7e4d5594873a617dc189f76a8a1","name":"citricsquid","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjQ4YTBhN2U0ZDU1OTQ4NzNhNjE3ZGMxODlmNzZhOGExIiwicHJvZmlsZU5hbWUiOiJjaXRyaWNzcXVpZCIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9lMWM2YzliNmRlODhmNDE4OGY5NzMyOTA5Yzc2ZGZjZDdiMTZhNDBhMDMxY2UxYjQ4NjhlNGQxZjg4OThlNGYifSwiQ0FQRSI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlL2MzYWY3ZmI4MjEyNTQ2NjQ1NThmMjgzNjExNThjYTczMzAzYzlhODVlOTZlNTI1MTEwMjk1OGQ3ZWQ2MGM0YTMifX19"}]}`,
		"00000000000000000000000000000003": `{"id":"00000000000000000000000000000003","name":"MalformedSession"`,
		"00000000000000000000000000000004": `{"id":"00000000000000000000000000000004","name":"NoTexture","properties":[]}`,
		"00000000000000000000000000000005": `{"id":"00000000000000000000000000000005","name":"MalformedTexProp","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA1IiwicHJvZmlsZU5hbWUiOiJNYWxmb3JtZWRUZXhQcm9wIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlL2NkOWNhNTVlOTg2MmYwMDNlYmZhMTg3MmE5MjQ0YWQ1ZjcyMWQ2YjllNjg4M2RkMWQ0MmY4N2RhZTEyNzY0OSJ9"}]}`,
		"00000000000000000000000000000008": `{"id":"00000000000000000000000000000008","name":"MalformedSTex","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA4IiwicHJvZmlsZU5hbWUiOiJNYWxmb3JtZWRTVGV4IiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlL01hbGZvcm1lZFRleHR1cmUifX19"}]}`,
		"00000000000000000000000000000009": `{"id":"00000000000000000000000000000009","name":"MalformedCTex","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA5IiwicHJvZmlsZU5hbWUiOiJNYWxmb3JtZWRDVGV4IiwidGV4dHVyZXMiOnsiU0tJTiI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlL2NkOWNhNTVlOTg2MmYwMDNlYmZhMTg3MmE5MjQ0YWQ1ZjcyMWQ2YjllNjg4M2RkMWQ0MmY4N2RhZTEyNzY0OSJ9LCJDQVBFIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvTWFsZm9ybWVkVGV4dHVyZSJ9fX0="}]}`,
		"00000000000000000000000000000010": `{"id":"00000000000000000000000000000010","name":"404STexture","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDEwIiwicHJvZmlsZU5hbWUiOiI0MDRTVGV4dHVyZSIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS80MDRUZXh0dXJlIn19fQ=="}]}`,
		"00000000000000000000000000000011": `{"id":"00000000000000000000000000000011","name":"404CTexture","properties":[{"name":"textures","value":"eyJ0aW1lc3RhbXAiOjAsInByb2ZpbGVJZCI6IjAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDExIiwicHJvZmlsZU5hbWUiOiI0MDRDVGV4dHVyZSIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9jZDljYTU1ZTk4NjJmMDAzZWJmYTE4NzJhOTI0NGFkNWY3MjFkNmI5ZTY4ODNkZDFkNDJmODdkYWUxMjc2NDkifSwiQ0FQRSI6eyJ1cmwiOiJodHRwOi8vdGV4dHVyZXMubWluZWNyYWZ0Lm5ldC90ZXh0dXJlLzQwNFRleHR1cmUifX19"}]}`,
		"00000000000000000000000000000014": `{"id":"00000000000000000000000000000014","properties":[]}`,
	}

	SessionProfilesSkinPath = map[string]string{
		"d9135e082f2244c89cb0bee234155292": "http://textures.minecraft.net/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649",
		"2f3665cc5e29439bbd14cb6d3a6313a7": "http://textures.minecraft.net/texture/b58a318a3ab7a0776378a28bb29e4287a85448abc3981a79f401e2b7ddf23",
		"5c115ca73efd41178213a0aff8ef11e0": "http://textures.minecraft.net/texture/6f736b4c3e2286cfad9b0d738fd7d9630d9e0a27721b7586e423cebce420da",
		"48a0a7e4d5594873a617dc189f76a8a1": "http://textures.minecraft.net/texture/e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f",
		"00000000000000000000000000000003": "",
		"00000000000000000000000000000004": "",
		"00000000000000000000000000000005": "",
		"00000000000000000000000000000008": "http://textures.minecraft.net/texture/MalformedTexture",
		"00000000000000000000000000000009": "http://textures.minecraft.net/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649",
		"00000000000000000000000000000010": "http://textures.minecraft.net/texture/404Texture",
		"00000000000000000000000000000011": "http://textures.minecraft.net/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649",
		"00000000000000000000000000000014": "",
	}

	Textures = map[string]string{
		// clone1018 skin
		"/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649": `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`,
		// citricquid skin
		"/texture/e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f": `iVBORw0KGgoAAAANSUhEUgAAAEAAAAAgCAIAAAAt/+nTAAAFJklEQVR42s2Y2W7UQBBF823kAYkdkh8ghH1JQsB2ZvH8KDwgsUPY9204VccyEWKMDGSYkdXqbld33aq6Vd2epaUZv09bW19v3JhOJm+vXPl8/TrPx83NN5cv80xHI1qGHzY2poMBktO6Xlq0H7gBN93ZAR9Yp+Pxl+3tmKyqaIfDbzdvYp6WYNLCGTAtS0C/u3oVoAERY4oCe4AObjpGBnuMz0JGoKqCJMNhtKMRHfxNHIJFVcUMwUHMmYUz4P21a+HvspTrEYTJJGhT10wG74uCIb6XRYtHoUyAcHaSBy5hkmEJM8bj4FJZBouGw0WkEIjBjQHBH0IxGllzojTt7OByBWgtTYsXgboOT6ebAyukHwwMiJlgNhOTYFdV/T+gWR9BGa61qEsMEAPRDLZoMpM1J9IjfR9iFCseaEZu0FYVu7kJdiK57waADCg8Tb1PlzOJ7hhCnroOoEVhuYwgiLiqAijJQL8sPTfMk+aMyyo8jwj8lI6634ITLeU/5wlRDD0QshaRGGAFccSkqswHhkG5PCuQ3H8DIEZdh/8AhCXjcUNxhnXt6Ys7m3I0GGBJvNXxyBcFAobRUABdq+gjue8GAK6tJ3KG1iAE6KIQja618Id3CUL6uFmVFKLT3KDKkm3ndMAFFE7Z9Fz0k98A/VFMcXYmdCM/GoEMM4Ao42WXLpdyplD0y3IeJy66w21ZZCQP4NoLBa/wazvTJHpeKwwUSyIHknvNGZdHOKvmksS/I3EQaTzW5d4swkgplKWpOROynkowZPRLG8l5nQkEwQMBF2bfyth9WofXIaFMa2tUspGtwuzJRMkOFvyyZPc2wLPMgtOeVh00kDbGEN2xPGsuJhkllwf6rLldV/e0vFmYNfpPDPiJvgxF1mVwVZnBfvR4VDeZQEDyruEh3VGRwlllSRvoJ5OmMP7BV154Is/XQJYfAF4uOu5LFlB0ew4Kojk3SJKWWsmNjlLu10VzT8kq0ta93/yenjnz8sKF15cuofjuoUMmpZ0X58+Dg7cM/S6joxosfHb2LGp219e9ODw+fZoHzxkQ5lkOGl69ungRGRSx1hKMTF+9Xby/d/jwk5Mn0Sq+J2trdJ6fO/fo+PGHx44BgqE3M3dEk6UTZAAFBChZbugienXNKyQxibdsCCYPdWSYV7KX3tkROHUKZSB4cPQocjhGixniY17dXl5miBrQ0AEQaGjRRGuOgoahn5d0eEDDgxcVYx/esrn1msm+emcacOvAAUPJjntDiW8M5Z2DB/eGkkl2RAevROm9GhtYyD4SicjgVIQNhW/BDWJd0FfvP6PQ3/ueJbz9ZxRCsaFUrg0lQ0OJL/eGEsV4CHkkveF56PBWS2g1DDEwmbW0Hs/M0EGmr96ZBiDt3csI0gcWHZzESibdiNaN0GT9wWf0rfrC8rpBB/TeOJCxwmqn3OBhh756uyhE3sBI3IYc+xpKlN0/coRMUoGMl0JAAbrFB1exBGHtke66Eyj+IclCcBgTELOVxaqX3pkGQEQ9ZNbrSzoQ0fg+PnHCyLK1ZAUfYrQG2rwUtAmAblprpUGw8iLm7Qjb+uqdaQByXoN3V1akuKajwLRTgVVfA0J4fd0bP8rAwaS3a+axpK05tFYkVXgIWqD66p1pQHBgdZWapaF+VQmUEPO4LyG2XOA8SS9btAR9Ega++op9/O/I/7R9a24401dv178SFm98qSd4JB+P/35KfRWwo1VFx6PeosS8wiz3cLDee3I7lHtAR11fvbPwfwejl0CqnPd41wAAAABJRU5ErkJggg==`,
		// citricquid cape
		"/texture/c3af7fb821254664558f28361158ca73303c9a85e96e5251102958d7ed60c4a3": `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`,
		// Malformed
		"/texture/MalformedTexture": `iVBORw0KGgoAAAANSUhEUgAAAEAAAAAgCAIAAAAt/+nTAAAFJklEQVR42s2Y2W7UQBBF823kAYkdkh8ghH1JQsB2ZvH8KDwgsUPY9204VccyEWKMDGSYkdXqbld33aq6Vd2epaUZv09bW19v3JhOJm+vXPl8/TrPx83NN5cv80xHI1qGHzY2poMBktO6Xlq0H7gBN93ZAR9Yp+Pxl+3tmKyqaIfDbzdvYp6WYNLCGTAtS0C/u3oVoAERY4oCe4AObjpGBnuMz0JGoKqCJMNhtKMRHfxNHIJFVcUMwUHMmYUz4P21a+HvspTrEYTJJGhT10wG74uCIb6XRYtHoUyAcHaSBy5hkmEJM8bj4FJZBouGw0WkEIjBjQHBH0IxGllzojTt7OByBWgtTYsXgboOT6ebAyukHwwMiJlgNhOTYFdV`,
		// Dud
		"/texture/404Texture": ``,
	}

	TexturesHash = map[string]string{
		"/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649": "a04a26d10218668a632e419ab073cf57",
		"/texture/e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f": "c05454f331fa93b3e38866a9ec52c467",
	}

}

// ReturnMux creates a mux for the mock Mojang API
func ReturnMux() *http.ServeMux {
	mux := http.NewServeMux()

	// APIProfile
	mux.HandleFunc("/users/profiles/minecraft/", func(w http.ResponseWriter, r *http.Request) {
		username := strings.TrimPrefix(r.URL.Path, "/users/profiles/minecraft/")
		username = strings.ToLower(username)
		if _, exists := APIProfiles[username]; exists {
			fmt.Fprintf(w, APIProfiles[username])
		} else {
			w.WriteHeader(204)
		}
	})

	// SessionProfile
	mux.HandleFunc("/session/minecraft/profile/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/session/minecraft/profile/")
		if _, exists := SessionProfiles[uuid]; exists {
			fmt.Fprintf(w, SessionProfiles[uuid])
		} else {
			w.WriteHeader(204)
		}
	})

	// Texture
	mux.HandleFunc("/texture/", func(w http.ResponseWriter, r *http.Request) {
		skinPath := r.URL.Path
		if _, exists := Textures[skinPath]; exists {
			textureBytes, _ := base64.StdEncoding.DecodeString(Textures[skinPath])
			w.Write(textureBytes)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "404 Not Found")
		}
	})

	// skins
	mux.HandleFunc("/skins/", func(w http.ResponseWriter, r *http.Request) {
		request := strings.TrimPrefix(r.URL.Path, "/skins/")

		if r.Host == "skins.example.net" {
			switch request {
			case "clone1018.png":
				textureBytes, _ := base64.StdEncoding.DecodeString(Textures["/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649"])
				w.Write(textureBytes)
				return
			case "citricsquid.png":
				textureBytes, _ := base64.StdEncoding.DecodeString(Textures["/texture/e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f"])
				w.Write(textureBytes)
				return
			case "MalformedTexture.png":
				textureBytes, _ := base64.StdEncoding.DecodeString(Textures["/texture/MalformedTexture"])
				w.Write(textureBytes)
				return
			}
		}
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 Not Found (%s)", r.Host)
	})

	// capes
	mux.HandleFunc("/capes/", func(w http.ResponseWriter, r *http.Request) {
		request := strings.TrimPrefix(r.URL.Path, "/capes/")

		if r.Host == "skins.example.net" {
			switch request {
			case "citricsquid.png":
				textureBytes, _ := base64.StdEncoding.DecodeString(Textures["/texture/c3af7fb821254664558f28361158ca73303c9a85e96e5251102958d7ed60c4a3"])
				w.Write(textureBytes)
				return
			}
		}
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 Not Found (%s)", r.Host)
	})

	mux.HandleFunc("/users/profiles/minecraft/ratelimitapi", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	mux.HandleFunc("/users/profiles/minecraft/RateLimitAPI", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	mux.HandleFunc("/users/profiles/minecraft/500api", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	mux.HandleFunc("/users/profiles/minecraft/500API", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	mux.HandleFunc("/session/minecraft/profile/00000000000000000000000000000001", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	mux.HandleFunc("/session/minecraft/profile/00000000000000000000000000000012", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	mux.HandleFunc("/session/minecraft/profile/00000000000000000000000000000013", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	mux.HandleFunc("/session/minecraft/profile/00000000000000000000000000000007", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	mux.HandleFunc("/texture/404Texture", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 Not Found")
	})

	mux.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "No route: %s\n\n", r.URL.Path)
		fmt.Fprintf(w, "Request: (%+v)\n\n", r)
	})

	return mux
}

// RewriteTransport is an http.RoundTripper that rewrites requests
// using the provided URL's Scheme and Host, and its Path as a prefix.
// The Opaque field is untouched.
// If Transport is nil, http.DefaultTransport is used
type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

// RoundTrip is used by the http.Client and rewrites the request to the testserver
func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// note that url.URL.ResolveReference doesn't work here
	// since t.u is an absolute url
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	//req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

// Setup returns the RewriteTransport for use in the http.Client
// It also returns the callback to close the server
func Setup(mux *http.ServeMux) (RewriteTransport, func()) {
	CreateMaps()
	testServer := httptest.NewServer(mux)

	TestURL = testServer.URL

	u, err := url.Parse(TestURL)
	if err != nil {
		log.Fatalln("failed to parse httptest.Server URL:", err)
	}

	return RewriteTransport{URL: u}, func() {
		testServer.Close()
	}
}
