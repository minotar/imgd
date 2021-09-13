package minecraft

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	// If we work with PNGs we need this
	_ "image/png"

	"github.com/pkg/errors"
)

type SessionProfileTextureProperty struct {
	TimestampMs uint64 `json:"timestamp"`
	ProfileUUID string `json:"profileId"`
	ProfileName string `json:"profileName"`
	IsPublic    bool   `json:"isPublic"`
	Textures    struct {
		Skin struct {
			Metadata struct {
				Model string `json:"model"`
			} `json:"metadata"`
			URL string `json:"url"`
		} `json:"SKIN"`
		Cape struct {
			URL string `json:"url"`
		} `json:"CAPE"`
	} `json:"textures"`
}

// DecodeTextureProperty takes a SessionProfileResponse and breaks it down into the Skin/Cape URLs for downloading them
func DecodeTextureProperty(sessionProfile SessionProfileResponse) (SessionProfileTextureProperty, error) {
	var texturesProperty *SessionProfileProperty
	for _, v := range sessionProfile.Properties {
		if v.Name == "textures" {
			texturesProperty = &v
			break
		}
	}

	if texturesProperty == nil {
		return SessionProfileTextureProperty{}, errors.New("unable to DecodeTextureProperty: no textures property")
	}

	profileTextureProperty := SessionProfileTextureProperty{}
	// Base64 decode the texturesProperty and further decode the JSON from it into profileTextureProperty
	err := json.NewDecoder(base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(texturesProperty.Value))).Decode(&profileTextureProperty)
	if err != nil {
		return SessionProfileTextureProperty{}, errors.Wrap(err, "unable to DecodeTextureProperty")
	}

	return profileTextureProperty, nil
}

func (mc *Minecraft) FetchTexturesWithSessionProfile(sessionProfile SessionProfileResponse) (User, Skin, Cape, error) {
	//  We have a sessionProfile!
	user := &User{UUID: sessionProfile.UUID, Username: sessionProfile.Username}
	skin := &Skin{Texture{Mc: mc}}
	cape := &Cape{Texture{Mc: mc}}

	profileTextureProperty, err := DecodeTextureProperty(sessionProfile)
	if err != nil {
		return *user, *skin, *cape, errors.Wrap(err, "failed to decode sessionProfile")
	}

	// We got oursleves a profileTextureProperty - now we can get a Skin/Cape

	err = skin.FetchWithTextureProperty(profileTextureProperty, "Skin")
	if err != nil {
		return *user, *skin, *cape, errors.Wrap(err, "not able to retrieve skin")
	}

	err = cape.FetchWithTextureProperty(profileTextureProperty, "Cape")
	if err != nil {
		return *user, *skin, *cape, errors.Wrap(err, "not able to retrieve cape")
	}
	return *user, *skin, *cape, nil
}
