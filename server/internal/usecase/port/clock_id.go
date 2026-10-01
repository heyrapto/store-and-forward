package port

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// Clock is an interface over time.Now() so use cases can be tested
// with a fixed, deterministic time.
type Clock interface {
	Now() time.Time
}

// RealClock is the production Clock implementation.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

// IDGenerator mints globally unique, k-sortable string IDs (ULIDs).
// Use cases call this instead of ulid.Make() directly so tests can
// inject a deterministic generator.
type IDGenerator interface {
	New() string
}

// ULIDGenerator is the production IDGenerator.
type ULIDGenerator struct{}

func (ULIDGenerator) New() string {
	return ulid.Make().String()
}
