package mcclient

import (
	"bytes"
	"compress/flate"
	"encoding/gob"
	"time"

	"github.com/minotar/minecraft"
	"google.golang.org/protobuf/proto"

	pb "github.com/minotar/imgd/pkg/mcclient/mcuser_proto"
)

type textures struct {
	SkinPath string
	//CapePath string
}

type McUser struct {
	Timestamp time.Time // Using a Seconds uint32 will be better for space
	minecraft.User
	Textures textures
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

func EncodeProtobufMcUser(user McUser) ([]byte, error) {
	u := &pb.McUserProto{
		Time:     uint32(user.Timestamp.Unix()),
		Username: user.Username,
		UUID:     user.UUID,
		SkinPath: user.Textures.SkinPath,
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
		User: minecraft.User{
			Username: u.Username,
			UUID:     u.UUID,
		},
		Textures: textures{
			SkinPath: u.SkinPath,
		},
	}

	return user, err
}

// ===

func EncodeGobMcUser(user McUser) ([]byte, error) {
	var bytes bytes.Buffer
	enc := gob.NewEncoder(&bytes)

	err := enc.Encode(user)
	if err != nil {
		return nil, err
	}

	return bytes.Bytes(), nil
}

func DecodeGobMcUser(gobBytes []byte) (McUser, error) {
	user := McUser{}

	reader := bytes.NewReader(gobBytes)
	dec := gob.NewDecoder(reader)
	err := dec.Decode(&user)
	if err != nil {
		return user, err
	}

	return user, err
}
