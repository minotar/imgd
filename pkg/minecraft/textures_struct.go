package minecraft

import (
	"context"
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
// Decoding the bytes into an image.Image is costly (as is re-encoding it back to a PNG)
type Texture struct {
	// texture image...
	Image image.Image
	// Mc is a pointer to the Minecraft struct that is then used for requests against the API
	Mc *Minecraft
	// Hash is the md5 of the image pixels
	Hash string
	// AlphaSig is a 4-byte signature of the background matte
	AlphaSig [4]uint8
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
		return err
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

func (t *Texture) loadTextureBody(texBody io.ReadCloser, err error) error {
	if err != nil {
		return errors.Wrap(err, "unable to Fetch Texture")
	}
	defer texBody.Close()
	return t.Decode(texBody)
}

// FetchWithTextureProperty takes a already decoded Texture Property and will request either Skin or Cape as instructed
func (t *Texture) FetchWithTextureProperty(sptp SessionProfileTextureProperty, texType textureType) error {
	err := t.loadTextureBody(t.Mc.TextureBodyFromTexturePropertyCtx(context.Background(), sptp, texType))
	if err != nil {
		return fmt.Errorf("FetchWithTextureProperty failed: %w", err)
	}
	return nil
}

// FetchWithSessionProfile will decode the Texture Property for you and request the Skin or Cape as instructed
// If requesting both Skin and Cape, this would result in 2 x decoding - use FetchWithTextureProperty instead
func (t *Texture) FetchWithSessionProfile(sessionProfile SessionProfileResponse, texType textureType) error {
	sptp, err := DecodeTextureProperty(sessionProfile)
	if err != nil {
		return errors.WithStack(err)
	}

	err = t.FetchWithTextureProperty(sptp, texType)
	if err != nil {
		return fmt.Errorf("FetchWithSessionProfile failed: %w", err)
	}
	return nil
}

// FetchWithUsername takes a username and will then request from UsernameAPI as specified in the Minecraft struct
func (t *Texture) FetchWithUsername(username string, texType textureType) error {
	err := t.loadTextureBody(t.Mc.TextureBodyFromUsernameCtx(context.Background(), username, texType))
	if err != nil {
		return fmt.Errorf("FetchWithUsername failed: %w", err)
	}
	return nil
}
