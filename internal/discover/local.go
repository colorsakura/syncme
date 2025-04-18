package discover

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/colorsakura/syncme/internal/gen"
	"github.com/colorsakura/syncme/internal/protocol"
	"github.com/syncthing/syncthing/lib/beacon"
	"github.com/syncthing/syncthing/lib/svcutil"
	"github.com/thejerf/suture/v4"
	"google.golang.org/protobuf/proto"
)

type localClient struct {
	*suture.Supervisor

	id       protocol.DeviceID
	addrList AddressLister
	name     string

	l *log.Logger

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

func NewLocal(uid protocol.DeviceID, addr string, addrList AddressLister, l *log.Logger) (FinderService, error) {
	c := &localClient{
		Supervisor:      suture.New("local", suture.Spec{}),
		l:               l,
		id:              uid,
		addrList:        addrList,
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

func (c *localClient) Lookup(_ context.Context, id protocol.DeviceID) (addresses []string, err error) {
	if cache, ok := c.Get(id); ok {
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
	pkg := &gen.Announce{
		Id: c.id[:],
	}
	bs, _ := proto.Marshal(pkg)

	if pktLen := 4 + len(bs); cap(msg) < pktLen {
		msg = make([]byte, 0, pktLen)
	}
	msg = msg[:4]
	binary.BigEndian.PutUint32(msg, Magic)
	msg = append(msg, bs...)
	for {
		c.l.Println("sendAnnouncements")
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
	b := c.beacon
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		buf, addr := b.Recv()
		if addr == nil {
			c.l.Fatal("recvAnnouncements: recv returned nil addr")
			continue
		}

		if len(buf) < 1 {
			c.l.Fatal("recvAnnouncements: recv returned too short buffer")
			continue
		}

		if magic := binary.BigEndian.Uint32(buf); magic != Magic {
			c.l.Printf("recvAnnouncements: recv %d bytes from %s: invalid magic", len(buf), addr)
			continue
		}

		var pkg gen.Announce
		err := proto.Unmarshal(buf, &pkg)
		if err != nil {
			c.l.Printf("recvAnnouncements: recv %d bytes from %s: %s", len(buf), addr, err)
			continue
		}

		c.Set(protocol.DeviceID(pkg.Id), CacheEntry{
			Addresses: pkg.Addresses,
			when:      time.Now(),
		})
		c.l.Printf("recvAnnouncements: recv %d bytes from %s: id=%x", len(buf), addr, pkg.Id)
	}
}
