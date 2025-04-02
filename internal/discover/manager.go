package discover

import (
	"sync"

	"github.com/thejerf/suture/v4"
)

// TODO:
type Manager interface{}

type manager struct {
	*suture.Supervisor
	uid     string
	cfg     []string
	finders map[string]Finder
	mut     sync.Mutex
}

func NewManager(uid string, cfg []string) Manager {
	m := &manager{
		uid:     uid,
		cfg:     cfg,
		finders: make(map[string]Finder),
		mut:     sync.Mutex{},
	}

	return m
}
