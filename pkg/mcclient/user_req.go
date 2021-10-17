package mcclient

import (
	"errors"

	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/util/log"
)

type UserReq minecraft.User

func (ur UserReq) GetUUID(logger log.Logger, mc *McClient) (newLogger log.Logger, uuid string, err error) {
	// If we were given a UUID, use it..!
	if ur.UUID != "" {
		return logger.With("uuid", ur.UUID), ur.UUID, nil
	}
	if ur.Username == "" {
		logger.Errorf("No UUID/Username was supplied: %v", ur)
		return logger, "", errors.New("no UUID/Username given")
	}

	// With given Username, get the UUID
	newLogger = logger.With("username", ur.Username)
	uuidEntry, err := mc.GetUUIDEntry(logger, ur.Username)
	if err != nil {
		return
	}
	uuid = uuidEntry.UUID
	newLogger = newLogger.With("uuid", uuid)
	return
}
