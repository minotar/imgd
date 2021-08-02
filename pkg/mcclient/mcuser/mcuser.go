package mcuser

import (
	"bytes"
	"compress/flate"
	"io/ioutil"

	"github.com/minotar/imgd/pkg/util/tinytime"
	"github.com/minotar/minecraft"
	"google.golang.org/protobuf/proto"
)

type McUser struct {
	minecraft.User
	Textures  textures
	Timestamp tinytime.TinyTime
	// Todo: what type to store User status?????
	Status uint8
}

// Decompress a Protobuf McUser
func DecompressMcUser(flatedBytes []byte) (McUser, error) {
	zr := flate.NewReader(bytes.NewReader(flatedBytes))
	protoBytes, err := ioutil.ReadAll(zr)
	if err != nil {
		return McUser{}, err
	}
	return decodeMcUserProtobuf(protoBytes)
}

// Decodes an McUser from the Protobuf
func decodeMcUserProtobuf(protoBytes []byte) (McUser, error) {
	pb := &McUserProto{}
	err := proto.Unmarshal(protoBytes, pb)

	user := McUser{
		Timestamp: tinytime.TinyTime(pb.Time),
		Status:    uint8(pb.Status),
		User: minecraft.User{
			Username: pb.Username,
			UUID:     pb.UUID,
		},
		Textures: textures{
			SkinPath: pb.SkinPath,
		},
	}

	// If the enum is set, then set True
	if pb.BaseURL == McUserProto_TEXTURES_MC_NET {
		user.Textures.TexturesMcNet = true
	}

	return user, err
}

func (u McUser) EncodeProtobuf() ([]byte, error) {
	pb := &McUserProto{
		Time:     uint32(u.Timestamp),
		Status:   McUserProto_UserStatus(u.Status),
		Username: u.Username,
		UUID:     u.UUID,
		SkinPath: u.Textures.SkinPath,
	}

	if u.Textures.TexturesMcNet {
		pb.BaseURL = McUserProto_TEXTURES_MC_NET
	}

	return proto.Marshal(pb)

}

func (u McUser) Compress() ([]byte, error) {
	var b bytes.Buffer
	zw, err := flate.NewWriter(&b, flate.BestCompression)
	if err != nil {
		return nil, err
	}

	protoBytes, err := u.EncodeProtobuf()
	if err != nil {
		return nil, err
	}

	zw.Write(protoBytes)
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}
