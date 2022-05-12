package minecraft

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	// If we work with PNGs we need this
	_ "image/png"
)

// textureType records the requirement of Skin vs. Cape
type textureType uint8

const (
	TextureSkin textureType = iota
	TextureCape
)

var (
	TextureType_name = map[textureType]string{
		TextureSkin: "SKIN",
		TextureCape: "CAPE",
	}
)

// DecodeTextureProperty decodes the Skin/Cape URLs from the SessionProfileResponse
func (spr SessionProfileResponse) DecodeTextureProperty() (SessionProfileTextureProperty, error) {
	var texturesProperty *SessionProfileProperty
	for _, v := range spr.Properties {
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
		return SessionProfileTextureProperty{}, fmt.Errorf("unable to DecodeTextureProperty: %w", err)
	}

	return profileTextureProperty, nil
}

// Deprecated in favour of struct method ?
// DecodeTextureProperty takes a SessionProfileResponse and breaks it down into the Skin/Cape URLs for downloading them
func DecodeTextureProperty(sessionProfile SessionProfileResponse) (SessionProfileTextureProperty, error) {
	return sessionProfile.DecodeTextureProperty()
}

type SessionProfileTextureProperty struct {
	Textures struct {
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
	ProfileUUID string `json:"profileId"`
	ProfileName string `json:"profileName"`
	TimestampMs uint64 `json:"timestamp"`
	IsPublic    bool   `json:"isPublic"`
}

// Remember to close the io.ReadCloser if the error is nil
func (mc *Minecraft) TextureBodyFromTexturePropertyCtx(ctx context.Context, sptp SessionProfileTextureProperty, texType textureType) (io.ReadCloser, error) {
	var url string
	if texType == TextureSkin {
		url = sptp.Textures.Skin.URL
	} else if texType == TextureCape {
		url = sptp.Textures.Cape.URL
	}

	if url == "" {
		return nil, fmt.Errorf("no URL for %s", TextureType_name[texType])
	}
	ctx = CtxWithSource(ctx, "TextureFetch")
	return mc.ApiRequestCtx(ctx, url)
}

func (mc *Minecraft) TextureBodyFromUsernameCtx(ctx context.Context, username string, texType textureType) (io.ReadCloser, error) {
	var baseURL string
	if texType == TextureSkin {
		baseURL = mc.Cfg.UsernameAPIConfig.SkinURL
	} else if texType == TextureCape {
		baseURL = mc.Cfg.UsernameAPIConfig.CapeURL
	}

	if baseURL == "" {
		return nil, fmt.Errorf("no UsernameAPI URL for %s", TextureType_name[texType])
	}
	ctx = CtxWithSource(ctx, "TextureFetch")
	return mc.ApiRequestCtx(ctx, baseURL+username+".png")
}

func (mc *Minecraft) FetchTexturesWithSessionProfile(sessionProfile SessionProfileResponse) (User, Skin, Cape, error) {
	//  We have a sessionProfile!
	user := User{UUID: sessionProfile.UUID, Username: sessionProfile.Username}
	skin := Skin{Texture{Mc: mc}}
	cape := Cape{Texture{Mc: mc}}

	profileTextureProperty, err := sessionProfile.DecodeTextureProperty()
	if err != nil {
		return user, skin, cape, fmt.Errorf("failed to decode sessionProfile: %w", err)
	}

	// We got oursleves a profileTextureProperty - now we can get a Skin/Cape

	err = skin.FetchWithTextureProperty(profileTextureProperty, TextureSkin)
	if err != nil {
		return user, skin, cape, fmt.Errorf("unable to retrieve skin: %w", err)
	}

	err = cape.FetchWithTextureProperty(profileTextureProperty, TextureCape)
	if err != nil {
		return user, skin, cape, fmt.Errorf("unable to retrieve cape: %w", err)
	}
	return user, skin, cape, nil
}
