package mcuser

import (
	"strings"

	"github.com/minotar/minecraft"
)

const TexturesBaseURL = "http://textures.minecraft.net/texture/"

type Textures struct {
	SkinPath string
	//SkinSlim bool (for "alex" support)
	//CapePath string

	// If TexturesMcNet is true, the SkinPath is just the part after the TexturesBaseURL
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

//
func (t Textures) Skin(mcClient *minecraft.Minecraft) (skin minecraft.Skin) {
	skin.URL = t.SkinURL()
	skin.Mc = mcClient
	return
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
