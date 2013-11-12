package minotar

////// THIS CODE IS RELEASED INTO THE PUBLIC DOMAIN. SEE "UNLICENSE" FOR MORE INFORMATION. http://unlicense.org //////

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
)

const (
	VALID_USERNAME_REGEX = `[a-zA-Z0-9_]+`
	SKIN_CACHE           = "skins/"
)

func FetchSkinFromLocal(path string) (Skin, error) {
	f, _ := os.Open(path)
	defer f.Close()

	return DecodeSkin(f)
}

func FetchSkinFromURL(url string) (Skin, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Skin{}, err
	}
	defer resp.Body.Close()

	return DecodeSkin(resp.Body)
}

func FetchSkinForUser(username string) (Skin, error) {
	path := "skins/" + username + ".png"

	if exists(path) == true {
		return FetchSkinFromLocal(path)
	} else {
		Skin, err := FetchSkinFromURL(URLForUser(username))

		saveAvatar(path, Skin)

		return Skin, err
	}
}

func URLForUser(username string) string {
	return "http://s3.amazonaws.com/MinecraftSkins/" + username + ".png"
}

func saveAvatar(path string, skin Skin) {
	f, _ := os.Create(path)
	defer f.Close()

	png.Encode(f, skin.Image)
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
