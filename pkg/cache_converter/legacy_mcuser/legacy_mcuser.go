package legacy_mcuser

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/minotar/imgd/pkg/minecraft"
)

type textures struct {
	SkinPath string
	//CapePath string
}

type mcUser struct {
	minecraft.User
	Textures  textures
	Timestamp time.Time
}

func NewMcUser(user minecraft.User, skinPath string, timestamp time.Time) *mcUser {
	return &mcUser{
		User: user,
		Textures: textures{
			SkinPath: skinPath,
		},
		Timestamp: timestamp,
	}
}

func DecodeMcUser(data []byte) (*mcUser, error) {
	user := &mcUser{}

	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	err := dec.Decode(user)
	if err != nil {
		return user, err
	}
	return user, nil

}

func (u *mcUser) Encode() ([]byte, error) {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	err := enc.Encode(u)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}
