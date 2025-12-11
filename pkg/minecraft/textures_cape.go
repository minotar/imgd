package minecraft

import (
	"errors"
	_ "image/png" // If we work with PNGs we need this
)

var (
	// ErrNoCape is returned when a player doesn't have a cape
	ErrNoCape = errors.New("player has no cape")
)

type Cape struct {
	Texture
}

// HasCape checks if a session profile contains a cape URL
func (c *Cape) HasCape(sessionProfile SessionProfileResponse) (bool, error) {
	sptp, err := DecodeTextureProperty(sessionProfile)
	if err != nil {
		return false, err
	}
	return sptp.Textures.Cape.URL != "", nil
}

func (mc *Minecraft) FetchCapeUUID(uuid string) (Cape, error) {
	cape := Cape{Texture{Mc: mc}}

	// Must be careful to not request same profile from session server more than once per ~30 seconds
	sessionProfile, err := mc.GetSessionProfile(uuid)
	if err != nil {
		return cape, err
	}

	// Check if the player has a cape before trying to fetch it
	hasCape, err := cape.HasCape(sessionProfile)
	if err != nil {
		return cape, err
	}
	if !hasCape {
		return cape, ErrNoCape
	}

	return cape, cape.FetchWithSessionProfile(sessionProfile, TextureCape)
}

func (mc *Minecraft) FetchCapeUsername(username string) (Cape, error) {
	cape := Cape{Texture{Mc: mc}}

	// First get the session profile to check if player has a cape
	// This is necessary to avoid making unnecessary texture requests
	uuid, err := mc.GetUUID(username)
	if err != nil {
		return cape, err
	}

	sessionProfile, err := mc.GetSessionProfile(uuid)
	if err != nil {
		return cape, err
	}

	// Check if the player has a cape before trying to fetch it
	hasCape, err := cape.HasCape(sessionProfile)
	if err != nil {
		return cape, err
	}
	if !hasCape {
		return cape, ErrNoCape
	}

	return cape, cape.FetchWithSessionProfile(sessionProfile, TextureCape)
}
