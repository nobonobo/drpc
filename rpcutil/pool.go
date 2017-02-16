package rpcutil

import (
	"fmt"
	"net/rpc"
	"sync"
)

type pcWrapper struct {
	mu sync.RWMutex
	*rpc.Client
	ret func(*rpc.Client) error
}

// Call ...
func (c *pcWrapper) Call(method string, args interface{}, reply interface{}) error {
	c.mu.RLock()
	err := c.Client.Call(method, args, reply)
	c.mu.RUnlock()
	if _, ok := err.(rpc.ServerError); !ok {
		c.mu.Lock()
		c.Client.Close()
		c.Client = nil
		c.mu.Unlock()
	}
	return err
}

// Go ...
func (c *pcWrapper) Go(method string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	done2 := make(chan *rpc.Call, 10)
	c.mu.RLock()
	res := c.Client.Go(method, args, reply, done2)
	c.mu.RUnlock()
	if _, ok := res.Error.(rpc.ServerError); !ok {
		c.mu.Lock()
		c.Client.Close()
		c.Client = nil
		c.mu.Unlock()
	}
	res.Done = done
	go func() {
		d, ok := <-done2
		if !ok {
			return
		}
		if _, ok := d.Error.(rpc.ServerError); !ok {
			c.mu.Lock()
			c.Client.Close()
			c.Client = nil
			c.mu.Unlock()
		}
		select {
		case done <- d:
		default:
		}
	}()
	return res
}

// Close ...
func (c *pcWrapper) Close() error {
	c.mu.RLock()
	retC := c.Client
	c.mu.RUnlock()
	return c.ret(retC)
}

// Pool ...
type Pool struct {
	sync.RWMutex
	pool    chan *rpc.Client
	factory func() (*rpc.Client, error)
	failed  int
}

// NewPool ...
func NewPool(max int, factory func() (*rpc.Client, error)) *Pool {
	p := &Pool{
		pool:    make(chan *rpc.Client, max),
		factory: factory,
	}
	for i := 0; i < max; i++ {
		p.pool <- nil
	}
	return p
}

// Get ...
func (p *Pool) Get() (Client, error) {
	c, ok := <-p.pool
	if !ok {
		return nil, fmt.Errorf("pool is closed")
	}
	if c == nil {
		newClient, err := p.factory()
		if err != nil {
			return nil, err
		}
		c = newClient
	}
	return &pcWrapper{
		Client: c,
		ret: func(retC *rpc.Client) error {
			if retC == nil {
				p.failed++
				return p.ret(nil)
			}
			p.failed = 0
			return p.ret(retC)
		},
	}, nil
}

func (p *Pool) get() chan<- *rpc.Client {
	p.RLock()
	ch := p.pool
	p.RUnlock()
	return ch
}

func (p *Pool) ret(c *rpc.Client) error {
	select {
	case p.pool <- c:
	default:
		// overflow
		if c != nil {
			return c.Close()
		}
	}
	return nil
}

// Close ...
func (p *Pool) Close() error {
	p.Lock()
	pool := p.pool
	p.pool = nil
	p.Unlock()
	close(pool)
	var firstErr error
	for c := range pool {
		if c != nil {
			if err := c.Close(); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	return firstErr
}

type svcWrapper struct {
	prefix string
	Client
}

// Call ...
func (c *svcWrapper) Call(method string, args interface{}, reply interface{}) error {
	return c.Client.Call(c.prefix+method, args, reply)
}

// Go ...
func (c *svcWrapper) Go(method string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return c.Client.Go(c.prefix+method, args, reply, done)
}

// GetService ...
func (p *Pool) GetService(prefix string) (Client, error) {
	c, err := p.Get()
	if err != nil {
		return nil, err
	}
	return &svcWrapper{prefix + ".", c}, nil
}
