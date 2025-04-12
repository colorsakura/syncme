package discover

import (
	"maps"
	"sync"
	"time"

	"github.com/colorsakura/syncme/internal/protocol"
	"github.com/thejerf/suture/v4"
)

// A cachedFinder is a Finder with associated cache timeouts.
type cachedFinder struct {
	Finder
	cacheTime    time.Duration
	negCacheTime time.Duration
	cache        *cache
	token        *suture.ServiceToken
}

// An error may implement cachedError, in which case it will be interrogated
// to see how long we should cache the error. This overrides the default
// negative cache time.
type cachedError interface {
	CacheFor() time.Duration
}

// A cache can be embedded wherever useful

type cache struct {
	entries map[protocol.DeviceID]CacheEntry
	mut     sync.Mutex
}

func newCache() *cache {
	return &cache{
		entries: make(map[protocol.DeviceID]CacheEntry),
	}
}

func (c *cache) Set(id protocol.DeviceID, ce CacheEntry) {
	c.mut.Lock()
	c.entries[id] = ce
	c.mut.Unlock()
}

func (c *cache) Get(id protocol.DeviceID) (CacheEntry, bool) {
	c.mut.Lock()
	ce, ok := c.entries[id]
	c.mut.Unlock()
	return ce, ok
}

func (c *cache) Cache() map[protocol.DeviceID]CacheEntry {
	c.mut.Lock()
	m := make(map[protocol.DeviceID]CacheEntry, len(c.entries))
	maps.Copy(m, c.entries)
	c.mut.Unlock()
	return m
}
