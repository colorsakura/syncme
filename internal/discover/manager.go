package discover

import (
	"context"
	"sync"
	"time"

	"github.com/thejerf/suture/v4"
)

// TODO:
type Manager interface {
	FinderService
}

type manager struct {
	*suture.Supervisor
	uid     string
	cfg     []string
	finders map[string]cachedFinder
	mut     sync.Mutex
}

func NewManager(uid string, cfg []string) Manager {
	m := &manager{
		uid:     uid,
		cfg:     cfg,
		finders: make(map[string]cachedFinder),
		mut:     sync.Mutex{},
	}

	return m
}

func (m *manager) serve(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (m *manager) setup(ctx context.Context) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	NewLocal("123456", "", []string{}, "ipv4")

	return nil
}

func (m *manager) Lookup(ctx context.Context, uid string) (addresses []string, err error) {
	m.mut.Lock()
	defer m.mut.Unlock()
	for _, finder := range m.finders {
		if addrs, err := finder.Lookup(ctx, uid); err == nil {
			addresses = append(addresses, addrs...)
			finder.cache.Set(uid, CacheEntry{
				Addresses: addrs,
				when:      time.Now(),
			})
		}
	}

	return
}

func (m *manager) Error() error {
	return nil
}

func (m *manager) String() string {
	return "manager"
}
