package mcuser

import (
	"io"
	"strings"

	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/util/log"
)

const TexturesBaseURL = "http://textures.minecraft.net/texture/"

type TextureIO struct {
	io.ReadCloser
	TextureID string
}

// DecodeTexture reads and closes the ReadCloser, returning a minecraft.Texture (and optional error)
func (tio TextureIO) DecodeTexture() (texture minecraft.Texture, err error) {
	defer tio.ReadCloser.Close()
	err = texture.Decode(tio.ReadCloser)

	//if err != nil {
	//	logger.Errorf("Failed to decode texture from %s: %v", mc.Caches.Textures.Name(), err)
	//	// Metrics stats Cache Decode Error
	//	return
	//}
	return
}

// MustDecodeSkin reads and closes the ReadCloser, returning a minecraft.Skin
func (tio TextureIO) MustDecodeSkin(logger log.Logger) (skin minecraft.Skin) {
	texture, err := tio.DecodeTexture()
	if err != nil {
		logger.Debugf("Falling back to Steve: %v", err)
		skin, _ = minecraft.FetchSkinForSteve()
		return
	}
	skin.Texture = texture
	return
}

func GetSteveTextureIO() TextureIO {
	// Todo: Can we optmize the Steve delivery - keep the bytes in memory and re-use?
	// is there a more efficient way to reuse the Steve bytes between requests (vs. a new buffer)?
	steve, _ := minecraft.GetSteveBytes()
	return TextureIO{
		ReadCloser: io.NopCloser(steve),
		TextureID:  minecraft.SteveHash,
	}

}

type Textures struct {
	// SkinPath changes based on whether the Texture's URL was prefixed by the TexturesBaseURL.
	// It will either be just the "hash" (part after the TexturesBaseURL) or a full URL
	SkinPath string
	//SkinSlim bool (for "alex" support)
	//CapePath string

	// TexturesMcNet is true when the SkinPath is just the part after the TexturesBaseURL
	// the Protobuf expresses this as an enum to support other values
	// This code does not need to support multiple values - unless new hosts are used
	TexturesMcNet bool
}

// Used to get a fully qualified URL for the Skin
func (t Textures) SkinURL() string {
	if t.TexturesMcNet {
		return TexturesBaseURL + t.SkinPath
	}
	return t.SkinPath
}

// After having made an API call, this can be used to create a textures object
func NewTexturesFromSessionProfile(sessionProfile minecraft.SessionProfileResponse) (t Textures, err error) {
	profileTextureProperty, err := minecraft.DecodeTextureProperty(sessionProfile)
	if err != nil {
		return
	}

	// If Skins URL starts with the known URL, set the "Path" to just the last bit
	if strings.HasPrefix(profileTextureProperty.Textures.Skin.URL, TexturesBaseURL) {
		t.TexturesMcNet = true
		t.SkinPath = strings.TrimPrefix(profileTextureProperty.Textures.Skin.URL, TexturesBaseURL)
	} else {
		t.TexturesMcNet = false
		t.SkinPath = profileTextureProperty.Textures.Skin.URL
	}

	// Other logic here for the Model / Slim skin, Capes etc.

	return t, nil
}
