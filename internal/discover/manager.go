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
	uid     protocol.DeviceID
	cfg     []string
	finders map[string]cachedFinder
	l       *log.Logger
	mut     sync.Mutex
}

func NewManager(uid protocol.DeviceID, cfg []string, l *log.Logger) Manager {
	m := &manager{
		Supervisor: suture.New("discover.manager", suture.Spec{}),
		uid:        uid,
		cfg:        cfg,
		finders:    make(map[string]cachedFinder),
		l:          l,
		mut:        sync.Mutex{},
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

	bcd, err := NewLocal(m.uid, ":18080", []string{}, m.l)
	if err != nil {
		return err
	}

	m.Add(bcd)
	return nil
}

func (m *manager) Lookup(ctx context.Context, uid protocol.DeviceID) (addresses []string, err error) {
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
	return "discover.manager"
}
