package uuid

import (
	"fmt"
	"time"

	"github.com/minotar/imgd/pkg/mcclient/status"
	"github.com/minotar/imgd/pkg/util/log"
	"github.com/minotar/imgd/pkg/util/tinytime"
	"github.com/minotar/minecraft"
)

// Todo: No inherit???
// type UUIDStatus status.Status

type UUIDEntry struct {
	UUID      string
	Timestamp tinytime.TinyTime
	Status    status.Status
}

func NewUUIDEntry(logger log.Logger, username string, uuid string, err error) UUIDEntry {
	return UUIDEntry{
		Status:    status.NewStatusFromError(logger, username, err),
		Timestamp: tinytime.NewTinyTime(time.Now()),
		UUID:      uuid,
	}
}

func (u UUIDEntry) IsValid() bool {
	// Todo: Status Okay or Stale?? - Ideally would want to log a failure of that check
	return minecraft.RegexUUIDPlain.MatchString(u.UUID)
}

// Todo: Does it make sense to check freshness of a Username:UUID mapping?
// It's only problematic when a Username switches between people
// If not... why even track Timestamp?
// Maybe debug info / headers / insight?
func (u UUIDEntry) IsFresh() bool {
	return true
}

func (u UUIDEntry) TTL() time.Duration {
	return u.Status.DurationUUID()
}

func (u UUIDEntry) String() string {
	if u.IsValid() {
		return fmt.Sprintf("{%s: %s}", u.UUID, u.Timestamp.Time())
	} else {
		return fmt.Sprintf("{%s: %s}", u.Status, u.Timestamp.Time())
	}
}

// Super simple format
//  |-----------------------------|
//  | status | timestamp | uuid   |
//  |-----------------------------|
//  | uint8  | uint32    | string |
//  |-----------------------------|

func (u UUIDEntry) Encode() []byte {
	// Status/uint8 takes 1 byte and TinyTime/uint32 takes 4
	buf := make([]byte, 5, 5+len(u.UUID))

	buf[0] = u.Status.Byte()
	u.Timestamp.Encode(buf[1:5])
	buf = append(buf, []byte(u.UUID)...)

	return buf
}

func DecodeUUIDEntry(buf []byte) UUIDEntry {
	return UUIDEntry{
		Status:    status.Status(buf[0]),
		Timestamp: tinytime.Decode(buf[1:5]),
		UUID:      string(buf[5:]),
	}
}
