package rpcutil

import "net/rpc"

// Client ...
type Client interface {
	Call(method string, args interface{}, reply interface{}) error
	Go(method string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call
	Close() error
}

// Clients ...
type Clients []Client

// Close ...
func (cl Clients) Close() error {
	var firstErr error
	for _, c := range cl {
		if err := c.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
