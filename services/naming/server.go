package naming

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
)

// RegisterInfo ...
type RegisterInfo struct {
	Addr     string
	Provides []string
}

// ServiceMap provide service name: array of host-addr
type ServiceMap map[string][]string

// Server ...
type Server struct {
	mu       sync.Mutex
	update   chan string
	monitor  *Monitor
	entries  map[string][]string
	modified bool
	buff     ServiceMap
}

// New ...
func New() *Server {
	ns := new(Server)
	ns.update = make(chan string)
	ns.monitor = NewMonitor(ns.update)
	ns.entries = map[string][]string{}
	ns.modified = true
	go func() {
		for dead := range ns.monitor.Dead() {
			ns.Remove(dead)
		}
	}()
	return ns
}

// Name ...
func (ns *Server) Name() string {
	return "NamingService"
}

// Rcvr ...
func (ns *Server) Rcvr() interface{} {
	return &NamingService{parent: ns}
}

// Update ...
func (ns *Server) Update(info RegisterInfo) {
	hash := md5.New()
	for _, v := range info.Provides {
		hash.Write([]byte(v))
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	ns.mu.Lock()
	defer ns.mu.Unlock()
	m, ok := ns.entries[info.Addr]
	if ok && len(m) > 0 && m[0] == digest {
		return
	}
	ns.entries[info.Addr] = append([]string{digest}, info.Provides...)
	ns.update <- info.Addr
	ns.modified = true
}

// Remove ...
func (ns *Server) Remove(addr string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if _, ok := ns.entries[addr]; ok {
		delete(ns.entries, addr)
		ns.monitor.Remove(addr)
		ns.modified = true
	}
}

// Data ...
func (ns *Server) Data() ServiceMap {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if !ns.modified {
		return ns.buff
	}
	res := ServiceMap{}
	for k, v := range ns.entries {
		for _, name := range v[1:] {
			res[name] = append(res[name], k)
		}
	}
	ns.buff = res
	ns.modified = false
	return res
}

// Close ...
func (ns *Server) Close() error {
	return ns.monitor.Close()
}

// NamingService ...
type NamingService struct {
	parent *Server
}

// Register ...
func (ns *NamingService) Register(info RegisterInfo, none *struct{}) error {
	ns.parent.Update(info)
	return nil
}

// Query ...
func (ns *NamingService) Query(name string, addrs *[]string) error {
	*addrs = ns.parent.Data()[name]
	return nil
}
