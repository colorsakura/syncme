package discover

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/syncthing/syncthing/lib/beacon"
	"github.com/syncthing/syncthing/lib/svcutil"
	"github.com/thejerf/suture/v4"
)

type localClient struct {
	*suture.Supervisor
	uid      string
	addrList []string
	name     string

	beacon          beacon.Interface
	localBcastStart time.Time
	localBcastTick  <-chan time.Time
	forceBcastTick  chan time.Time

	*cache
}

const (
	BroadcastInterval = 30 * time.Second
	CacheLifeTime     = 3 * BroadcastInterval
	Magic             = uint32(0x2EA7D90B) // same as in BEP
)

func NewLocal(uid string, addr string, addrList []string, name string) (FinderService, error) {
	c := &localClient{
		Supervisor:      suture.New("local", suture.Spec{}),
		uid:             uid,
		addrList:        addrList,
		name:            name,
		localBcastTick:  time.NewTicker(BroadcastInterval).C,
		forceBcastTick:  make(chan time.Time),
		localBcastStart: time.Now(),
		cache:           newCache(),
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if host == "" {
		bcPort, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}

		c.beacon = beacon.NewBroadcast(bcPort)
	}

	c.Add(c.beacon)
	c.Add(svcutil.AsService(c.recvAnnouncements, fmt.Sprintf("%s/recv", c)))
	c.Add(svcutil.AsService(c.sendAnnouncements, fmt.Sprintf("%s/send", c)))

	return c, nil
}

func (c *localClient) Lookup(_ context.Context, uid string) (addresses []string, err error) {
	if cache, ok := c.Get(uid); ok {
		if time.Since(cache.when) < CacheLifeTime {
			addresses = cache.Addresses
		}
	}
	return
}

func (c *localClient) Error() error {
	return c.beacon.Error()
}

func (c *localClient) String() string {
	return c.name
}

func (c *localClient) sendAnnouncements(ctx context.Context) error {
	var msg []byte
	for {
		c.beacon.Send(msg)

		select {
		case <-c.localBcastTick:
		case <-c.forceBcastTick:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *localClient) recvAnnouncements(ctx context.Context) error {
	return nil
}
