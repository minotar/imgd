package minecraft

import _ "image/png" // If we work with PNGs we need this

type Cape struct {
	Texture
}

func (mc *Minecraft) FetchCapeUUID(uuid string) (Cape, error) {
	cape := &Cape{Texture{Mc: mc}}

	// Must be careful to not request same profile from session server more than once per ~30 seconds
	sessionProfile, err := mc.GetSessionProfile(uuid)
	if err != nil {
		return *cape, err
	}

	return *cape, cape.FetchWithSessionProfile(sessionProfile, "Cape")
}

func (mc *Minecraft) FetchCapeUsername(username string) (Cape, error) {
	cape := &Cape{Texture{Mc: mc}}

	return *cape, cape.FetchWithUsername(username, "Cape")
}
