package mcclient

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
	"strings"
	"time"

	"github.com/minotar/minecraft"
	"google.golang.org/protobuf/proto"

	pb "github.com/minotar/imgd/pkg/mcclient/mcuser_proto"
)

const TexturesBaseURL = "http://textures.minecraft.net/texture/"

// Todo: Does it make more sense to combine this with mcuser_proto over there?

// Todo: Does it make sense to make a TinyTime type?

// Todo: should I rename this towards Marshal / UnMarshal ?
// Should I be accepting an existing interface vs. returning

type textures struct {
	SkinPath string
	//SkinSlim bool (for "alex" support)
	//CapePath string

	// If TexturesMcNet is true, the SkinPath is just the part after the TexturesBaseURL
	// the Protobuf expresses this as an enum to support other values
	// This code does not need to support multiple values - unless new hosts are used
	TexturesMcNet bool
}

// Used to get a fully qualified URL for the Skin
func (t *textures) SkinURL() string {
	if t.TexturesMcNet {
		return TexturesBaseURL + t.SkinPath
	}
	return t.SkinPath
}

// After having made an API call, this can be used to create a textures object
func NewTexturesFromSessionProfile(sessionProfile minecraft.SessionProfileResponse) (textures, error) {
	var t textures
	profileTextureProperty, err := minecraft.DecodeTextureProperty(sessionProfile)
	if err != nil {
		return t, err
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

type McUser struct {
	Timestamp time.Time // Using a Seconds uint32 will be better for space
	minecraft.User
	Textures textures
	// Todo: what type to store User status?????
	Status uint8
}

func PackMcUser(user McUser) ([]byte, error) {
	var b bytes.Buffer
	zw, err := flate.NewWriter(&b, flate.BestCompression)
	if err != nil {
		return nil, err
	}

	protoBytes, err := EncodeProtobufMcUser(user)
	if err != nil {
		return nil, err
	}

	zw.Write(protoBytes)
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func UnPackUser(packedBytes []byte) (McUser, error) {
	zr := flate.NewReader(bytes.NewReader(packedBytes))
	protoBytes, err := ioutil.ReadAll(zr)
	if err != nil {
		return McUser{}, err
	}
	return DecodeProtobufMcUser(protoBytes)

}

func EncodeProtobufMcUser(user McUser) ([]byte, error) {
	u := &pb.McUserProto{
		Time:     uint32(user.Timestamp.Unix()),
		Status:   pb.McUserProto_UserStatus(user.Status),
		Username: user.Username,
		UUID:     user.UUID,
		SkinPath: user.Textures.SkinPath,
	}

	if user.Textures.TexturesMcNet {
		u.BaseURL = pb.McUserProto_TEXTURES_MC_NET
	}

	return proto.Marshal(u)
}

func getTimeFromEpoch32(expirySeconds uint32) (t time.Time) {
	return time.Unix(int64(expirySeconds), 0).UTC()
}

func DecodeProtobufMcUser(protoBytes []byte) (McUser, error) {
	u := &pb.McUserProto{}
	err := proto.Unmarshal(protoBytes, u)

	user := McUser{
		Timestamp: getTimeFromEpoch32(u.Time),
		Status:    uint8(u.Status),
		User: minecraft.User{
			Username: u.Username,
			UUID:     u.UUID,
		},
		Textures: textures{
			SkinPath: u.SkinPath,
		},
	}

	// If the enum is set, then set True
	if u.BaseURL == pb.McUserProto_TEXTURES_MC_NET {
		user.Textures.TexturesMcNet = true
	}

	return user, err
}
