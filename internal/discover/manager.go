package discover

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/colorsakura/syncme/internal/protocol"
	"github.com/syncthing/syncthing/lib/svcutil"
	"github.com/thejerf/suture/v4"
)

// TODO:
type Manager interface {
	FinderService
}

type manager struct {
	*suture.Supervisor
	id            protocol.DeviceID
	cfg           []string
	addressLister AddressLister
	finders       map[string]cachedFinder

	l   *log.Logger
	mut sync.Mutex
}

func NewManager(id protocol.DeviceID, cfg []string, lister AddressLister, l *log.Logger) Manager {
	m := &manager{
		Supervisor:    suture.New("discover.manager", suture.Spec{}),
		id:            id,
		cfg:           cfg,
		addressLister: lister,
		finders:       make(map[string]cachedFinder),
		l:             l,
		mut:           sync.Mutex{},
	}
	m.Add(svcutil.AsService(m.serve, m.String()))
	return m
}

func (m *manager) serve(ctx context.Context) error {
	m.Setup()
	<-ctx.Done()
	return nil
}

func (m *manager) Setup() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	bcd, err := NewLocal(m.id, ":18080", m.addressLister, m.l)
	if err != nil {
		return err
	}

	m.Add(bcd)
	return nil
}

func (m *manager) Lookup(ctx context.Context, id protocol.DeviceID) (addresses []string, err error) {
	m.mut.Lock()
	defer m.mut.Unlock()
	for _, finder := range m.finders {
		if addrs, err := finder.Lookup(ctx, id); err == nil {
			addresses = append(addresses, addrs...)
			finder.cache.Set(id, CacheEntry{
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
	return "discover.manager"
}
