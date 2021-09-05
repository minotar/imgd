package mcuser

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/minotar/imgd/pkg/util/log"

	"github.com/minotar/imgd/pkg/mcclient/status"
	"github.com/minotar/imgd/pkg/util/tinytime"
	"github.com/minotar/minecraft"
	"google.golang.org/protobuf/proto"
)

type McUser struct {
	minecraft.User
	Textures  textures
	Timestamp tinytime.TinyTime
	Status    status.Status
}

func NewMcUser(logger log.Logger, uuid string, sessionProfile minecraft.SessionProfileResponse, err error) McUser {
	// Todo: handle this error!
	textures, _ := NewTexturesFromSessionProfile(sessionProfile)
	return McUser{
		Status:    status.NewStatusFromError(logger, uuid, err),
		Timestamp: tinytime.NewTinyTime(time.Now()),
		User:      sessionProfile.User,
		Textures:  textures,
	}
}

func (u McUser) IsValid() bool {
	// Todo: Status Okay or Stale?? - Ideally would want to log a failure of that check
	return u.Textures.SkinPath != ""
}

func (u McUser) IsFresh() bool {
	// Add the Timestamp to the Fresh TTL to get the point it's no longer fresh
	staleTime := u.Timestamp.Time().Add(status.UserFreshTTL)
	return time.Now().Before(staleTime)
}

func (u McUser) TTL() time.Duration {
	return u.Status.DurationUser()
}

func (u McUser) String() string {
	if u.IsValid() {
		return fmt.Sprintf("{%s:%s  %s}", u.Username, u.UUID, u.Timestamp.Time())
	} else {
		return fmt.Sprintf("{%s %s:%s %s}", u.Status, u.Username, u.UUID, u.Timestamp.Time())
	}
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
		Status:    status.Status(pb.Status),
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
