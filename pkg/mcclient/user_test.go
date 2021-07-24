package mcclient

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/minotar/minecraft"
)

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

func TestEncodeDecodeMcUser(t *testing.T) {
	user := McUser{
		Timestamp: time.Now(),
		User: minecraft.User{
			Username: "LukeHandle",
			UUID:     "5c115ca73efd41178213a0aff8ef11e0",
		},
		Textures: textures{
			SkinPath:      "6f736b4c3e2286cfad9b0d738fd7d9630d9e0a27721b7586e423cebce420da",
			TexturesMcNet: true,
		},
	}

	gobBytes, err := EncodeGobMcUser(user)
	if err != nil {
		t.Fatalf("Gob Encode failed with: %s", err)
	}
	fmt.Printf("The Gob output was length %d: %+v\n", len(gobBytes), gobBytes)

	protoBytes, err := EncodeProtobufMcUser(user)
	if err != nil {
		t.Fatalf("Protobuf Encode failed with: %s", err)
	}
	fmt.Printf("The Protobuf output was length %d: %+v\n", len(protoBytes), protoBytes)

	gobUser, err := DecodeGobMcUser(gobBytes)
	if err != nil {
		t.Fatalf("Gob Decode failed with: %s", err)
	}

	protoUser, err := DecodeProtobufMcUser(protoBytes)
	if err != nil {
		t.Fatalf("Protobuf Decode failed with: %s", err)
	}

	if gobUser.Username != protoUser.Username {
		t.Errorf("Gob Username \"%s\" vs. Protobuf Username \"%s\"", gobUser.Username, protoUser.Username)
	}
}

func TestPackUnPackMcUser(t *testing.T) {
	user := McUser{
		Timestamp: time.Now(),
		User: minecraft.User{
			Username: "LukeHandle",
			UUID:     "5c115ca73efd41178213a0aff8ef11e0",
		},
		Textures: textures{
			SkinPath:      "6f736b4c3e2286cfad9b0d738fd7d9630d9e0a27721b7586e423cebce420da",
			TexturesMcNet: true,
		},
	}

	packedBytes, err := PackMcUser(user)
	if err != nil {
		t.Fatalf("Flated Protobuf Encode failed with: %s", err)
	}
	fmt.Printf("The Flated Protobuf output was length %d: %+v\n", len(packedBytes), packedBytes)

	packedUser, err := UnPackUser(packedBytes)
	if err != nil {
		t.Fatalf("Flated Protobuf Decode failed with: %s", err)
	}

	if user.Username != packedUser.Username {
		t.Errorf("Original Username \"%s\" vs. Fkated/Protobuf Username \"%s\"", user.Username, packedUser.Username)
	}
}
