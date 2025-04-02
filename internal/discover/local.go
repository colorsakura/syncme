package discover

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/syncthing/syncthing/lib/beacon"
	"github.com/thejerf/suture/v4"
)

type localClient struct {
	*suture.Supervisor
	uid      string
	addrList []string
	name     string

	beacon         beacon.Interface
	localBcastTime time.Time
	localBcastTick <-chan time.Time
	forceBcastTick chan time.Time
}

const (
	BroadcastInterval = 30 * time.Second
	Magic             = uint32(0x2EA7D90B) // same as in BEP
)

func NewLocal(uid string, addr string, addrList []string, name string) (FinderService, error) {
	c := &localClient{
		Supervisor:     suture.New("local", suture.Spec{}),
		uid:            uid,
		addrList:       addrList,
		name:           name,
		localBcastTick: time.NewTicker(BroadcastInterval).C,
		forceBcastTick: make(chan time.Time),
		localBcastTime: time.Now(),
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

	return c, nil
}

func (c *localClient) Lookup(_ context.Context, uid string) (address []string, err error) {
	// TODO:
	return c.addrList, nil
}

func (c *localClient) Error() error {
	return c.beacon.Error()
}

func (c *localClient) String() string {
	return c.name
}
