package minecraft

import _ "image/png" // If we work with PNGs we need this

type Skin struct {
	Texture
}

func (mc *Minecraft) FetchSkinUUID(uuid string) (Skin, error) {
	skin := &Skin{Texture{Mc: mc}}

	// Must be careful to not request same profile from session server more than once per ~30 seconds
	sessionProfile, err := mc.GetSessionProfile(uuid)
	if err != nil {
		return *skin, err
	}

	return *skin, skin.FetchWithSessionProfile(sessionProfile, "Skin")
}

func (mc *Minecraft) FetchSkinUsername(username string) (Skin, error) {
	skin := &Skin{Texture{Mc: mc}}

	return *skin, skin.FetchWithUsername(username, "Skin")
}
