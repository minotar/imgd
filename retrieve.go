package minotar

import (
	"net/http"
)

const (
	VALID_USERNAME_REGEX = `[a-zA-Z0-9_]+`
    SKIN_CACHE = 'skins/'
)

func FetchSkinFromURL(url string) (Skin, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Skin{}, err
	}
	defer resp.Body.Close()

	return DecodeSkin(resp.Body)
}

func FetchSkinForUser(username string) (Skin, error) {
	return FetchSkinFromURL(URLForUser(username))
}

func URLForUser(username string) string {
	return "http://s3.amazonaws.com/MinecraftSkins/" + username + ".png"
}

func HasLocalFile(username string) bool {
    if _, err := os.Stat("./"+SKIN_CACHE+"/"+username+".png"); err != nil {
        if os.IsNotExist(err) {
            return true
        }
    }
    return false
}