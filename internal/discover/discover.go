package discover

import (
	"context"

	"github.com/thejerf/suture/v4"
)

type Finder interface {
	Lookup(ctx context.Context, uid string) (address []string, err error)
	Error() error
	String() string
}

type FinderService interface {
	Finder
	suture.Service
}
