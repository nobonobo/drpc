package rpcutil

import (
	"net/rpc"
	"sync"
)

// ClientFactory for NewClient
type ClientFactory func(addr string) (*rpc.Client, error)

// Peers ...
type Peers struct {
	sync.RWMutex
	maxConns int
	factory  ClientFactory
	m        map[string]*Pool
}

// NewPeers ...
func NewPeers(maxConns int, factory ClientFactory) *Peers {
	return &Peers{
		maxConns: maxConns,
		factory:  factory,
		m:        make(map[string]*Pool),
	}
}

// Join ...
func (p *Peers) Join(addr string) error {
	p.Lock()
	defer p.Unlock()
	pool := p.m[addr]
	if pool != nil {
		if err := pool.Close(); err != nil {
			return err
		}
	}
	p.m[addr] = NewPool(
		p.maxConns,
		func() (*rpc.Client, error) { return p.factory(addr) },
	)
	return nil
}

// Leave ...
func (p *Peers) Leave(addr string) error {
	p.Lock()
	defer p.Unlock()
	pool := p.m[addr]
	delete(p.m, addr)
	if pool != nil {
		return pool.Close()
	}
	return nil
}

// Get ...
func (p *Peers) Get(addr string) *Pool {
	p.RLock()
	defer p.RUnlock()
	return p.m[addr]
}

// List ...
func (p *Peers) List() []string {
	p.RLock()
	defer p.RUnlock()
	res := make([]string, 0, len(p.m))
	for k := range p.m {
		res = append(res, k)
	}
	return res
}

// LeaveAll ...
func (p *Peers) LeaveAll() error {
	p.Lock()
	defer p.Unlock()
	var firstErr error
	for _, pool := range p.m {
		if pool != nil {
			if err := pool.Close(); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	p.m = make(map[string]*Pool)
	return firstErr
}
