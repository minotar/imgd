package mcclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/minotar/minecraft"
)

func TestEncodeDecodeMcUser(t *testing.T) {
	user := McUser{
		Timestamp: time.Now(),
		User: minecraft.User{
			Username: "LukeHandle",
			UUID:     "5c115ca73efd41178213a0aff8ef11e0",
		},
		Textures: textures{
			SkinURL: "http://textures.minecraft.net/texture/6f736b4c3e2286cfad9b0d738fd7d9630d9e0a27721b7586e423cebce420da",
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

	packedBytes, err := PackMcUser(user)
	if err != nil {
		t.Fatalf("Flated Protobuf Encode failed with: %s", err)
	}
	fmt.Printf("The Flated Protobuf output was length %d: %+v\n", len(packedBytes), packedBytes)

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
