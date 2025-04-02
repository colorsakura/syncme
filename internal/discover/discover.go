package discover

import (
	"context"
	"time"

	"github.com/thejerf/suture/v4"
)

type Finder interface {
	Lookup(ctx context.Context, uid string) (address []string, err error)
	Error() error
	String() string
}

type CacheEntry struct {
	Addresses []string
	when      time.Time
}

type FinderService interface {
	Finder
	suture.Service
}
