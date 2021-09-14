package minecraft

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/draw"
	"io"

	// If we work with PNGs we need this
	_ "image/png"

	"github.com/pkg/errors"
)

// Texture is our structure for the Cape/Skin structs and the functions for dealing with it
type Texture struct {
	// texture image...
	Image image.Image
	// md5 hash of the texture image
	Hash string
	// Location we grabbed the texture from. Mojang/S3/Char
	Source string
	// 4-byte signature of the background matte for the texture
	AlphaSig [4]uint8
	// URL of the texture
	URL string
	// M is a pointer to the Minecraft struct that is then used for requests against the API
	Mc *Minecraft
}

// CastToNRGBA takes image bytes and converts to NRGBA format if needed
func (t *Texture) CastToNRGBA(r io.Reader) error {
	// Decode the skin
	textureImg, format, err := image.Decode(r)
	if err != nil {
		return errors.Wrap(err, "unable to CastToNRGBA")
	}

	// Convert it to NRGBA if necessary
	textureFinal := textureImg
	if format != "NRGBA" {
		bounds := textureImg.Bounds()
		textureFinal = image.NewNRGBA(bounds)
		draw.Draw(textureFinal.(draw.Image), bounds, textureImg, image.Pt(0, 0), draw.Src)
	}

	t.Image = textureFinal
	return nil
}

// Decode takes the image bytes and turns it into our Texture struct
func (t *Texture) Decode(r io.Reader) error {
	err := t.CastToNRGBA(r)
	if err != nil {
		return errors.WithStack(err)
	}

	// And md5 hash its pixels
	hasher := md5.New()
	hasher.Write(t.Image.(*image.NRGBA).Pix)
	t.Hash = fmt.Sprintf("%x", hasher.Sum(nil))

	// Create the alpha signature
	img := t.Image.(*image.NRGBA)
	t.AlphaSig = [...]uint8{
		img.Pix[0],
		img.Pix[1],
		img.Pix[2],
		img.Pix[3],
	}

	return nil
}

// Fetch performs the GET for the texture, doing any required conversion and saving our Image property
func (t *Texture) Fetch() error {
	apiBody, err := t.Mc.apiRequest(t.URL)
	if apiBody != nil {
		defer apiBody.Close()
	}
	if err != nil {
		return errors.Wrap(err, "unable to Fetch Texture")
	}

	err = t.Decode(apiBody)
	if err != nil {
		return errors.Wrap(err, "unable to Decode Texture")
	}
	return nil
}

// FetchWithTextureProperty takes a already decoded Texture Property and will request either Skin or Cape as instructed
func (t *Texture) FetchWithTextureProperty(profileTextureProperty SessionProfileTextureProperty, textureType string) error {
	if textureType == "Skin" {
		t.URL = profileTextureProperty.Textures.Skin.URL
	} else if textureType == "Cape" {
		t.URL = profileTextureProperty.Textures.Cape.URL
	} else {
		return errors.New("Unknown textureType")
	}

	if t.URL == "" {
		return errors.Errorf("%s URL not present", textureType)
	}
	t.Source = "SessionProfile"

	err := t.Fetch()
	if err != nil {
		return errors.Wrap(err, "FetchWithTextureProperty failed")
	}
	return nil
}

// FetchWithSessionProfile will decode the Texture Property for you and request the Skin or Cape as instructed
// If requesting both Skin and Cape, this would result in 2 x decoding - use FetchWithTextureProperty instead
func (t *Texture) FetchWithSessionProfile(sessionProfile SessionProfileResponse, textureType string) error {
	profileTextureProperty, err := DecodeTextureProperty(sessionProfile)
	if err != nil {
		return errors.WithStack(err)
	}

	err = t.FetchWithTextureProperty(profileTextureProperty, textureType)
	if err != nil {
		return errors.Wrap(err, "FetchWithSessionProfile failed")
	}
	return nil
}

// FetchWithUsername takes a username and will then request from UsernameAPI as specified in the Minecraft struct
func (t *Texture) FetchWithUsername(username string, textureType string) error {
	if textureType == "Skin" && t.Mc.Cfg.UsernameAPIConfig.SkinURL != "" {
		t.URL = t.Mc.Cfg.UsernameAPIConfig.SkinURL + username + ".png"
	} else if textureType == "Cape" && t.Mc.Cfg.UsernameAPIConfig.CapeURL != "" {
		t.URL = t.Mc.Cfg.UsernameAPIConfig.CapeURL + username + ".png"
	} else {
		return errors.New("Unkown textureType or missing UsernameAPI lookup URL")
	}
	t.Source = "UsernameAPI"

	err := t.Fetch()
	if err != nil {
		return errors.Wrap(err, "FetchWithUsername failed")
	}
	return nil
}
