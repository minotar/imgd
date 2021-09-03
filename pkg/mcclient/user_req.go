package mcclient

import (
	"context"
	"errors"

	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/minecraft"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type UserReq minecraft.User

func (ur UserReq) GetUUID(logger log.Logger, ctx context.Context, mc *McClient) (newLogger log.Logger, uuid string, err error) {
	// If we were given a UUID, use it..!
	if ur.UUID != "" {
		return logger.With("uuid", ur.UUID), ur.UUID, nil
	}
	if ur.Username == "" {
		logger.Errorf("No UUID/Username was supplied: %v", ur)
		return logger, "", errors.New("no UUID/Username given")
	}

	tracer := otel.Tracer("github.com/minotar/imgd/pkg/mcclient")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GetUUIDEntry")
	defer span.End()

	// With given Username, get the UUID
	newLogger = logger.With("username", ur.Username)
	uuidEntry, err := mc.GetUUIDEntry(logger, ctx, ur.Username)
	if err != nil {
		return
	}
	uuid = uuidEntry.UUID
	newLogger = newLogger.With("uuid", uuid)
	return
}
