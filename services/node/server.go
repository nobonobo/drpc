package node

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"golang.org/x/net/websocket"

	"time"

	"github.com/nobonobo/drpc"
	"github.com/nobonobo/drpc/rpcutil"
	"github.com/nobonobo/drpc/services"
	"github.com/nobonobo/drpc/services/naming"
)

// Server ...
type Server struct {
	*rpc.Server
	done        chan struct{}
	act         chan struct{}
	registory   *services.Registory
	selfAddr    string
	peers       *rpcutil.Peers
	httpHandler http.Handler
	nsFactory   func() (rpcutil.Client, error)
}

// New ...
func New(selfAddr string) (*Server, error) {
	s := new(Server)
	factory := func(addr string) (*rpc.Client, error) {
		if selfAddr == addr {
			return drpc.DefaultLocalFactory(s)
		}
		return drpc.DefaultRemoteFactory(addr)
	}
	s.Server = rpc.NewServer()
	s.done = make(chan struct{})
	s.act = make(chan struct{})
	s.registory = services.NewRegistory()
	s.selfAddr = selfAddr
	s.peers = rpcutil.NewPeers(drpc.DefaultMaxConns, factory)
	s.httpHandler = websocket.Handler(s.handle)
	if err := s.Join(s.SelfAddr()); err != nil {
		return nil, err
	}
	if err := s.Register(s); err != nil {
		return nil, err
	}
	go func() {
		tm := time.NewTimer(drpc.DefaultHertbeatInterval)
		failed := 0
		action := func() {
			if err := s.activate(); err != nil {
				failed++
				log.Println("activate failed:", err)
			} else {
				failed = 0
			}
			d := drpc.DefaultHertbeatInterval * time.Duration(failed+1)
			if d > drpc.DefaultHertbeatMax {
				d = drpc.DefaultHertbeatMax
			}
			tm.Reset(d)
		}
		for {
			select {
			case <-s.done:
				return
			case <-s.act:
				action()
			case <-tm.C:
				action()
			}
		}
	}()
	return s, nil
}

// Register ...
func (s *Server) Register(svc services.Service) error {
	err := s.Server.RegisterName(svc.Name(), svc.Rcvr())
	if err != nil {
		return err
	}
	return s.registory.Register(svc)
}

// RegisterName ...
func (s *Server) RegisterName(name string, svc services.Service) error {
	err := s.Server.RegisterName(name, svc.Rcvr())
	if err != nil {
		return err
	}
	return s.registory.RegisterName(name, svc)
}

// SelfAddr ...
func (s *Server) SelfAddr() string {
	return s.selfAddr
}

// Provides ...
func (s *Server) Provides() []string {
	return s.registory.List()
}

// Join ...
func (s *Server) Join(addr string) error {
	if s.peers.Get(addr) != nil {
		return nil
	}
	return s.peers.Join(addr)
}

// Leave ...
func (s *Server) Leave(addr string) error {
	return s.peers.Leave(addr)
}

// LeaveAll ...
func (s *Server) LeaveAll() error {
	return s.peers.LeaveAll()
}

// Peers ...
func (s *Server) Peers() []string {
	return s.peers.List()
}

// Get ...
func (s *Server) Get(addr, service string) (rpcutil.Client, error) {
	p := s.peers.Get(addr)
	if p == nil {
		return nil, fmt.Errorf("not found node: %s", addr)
	}
	return p.GetService(service)
}

// Close ...
func (s *Server) Close() error {
	close(s.done)
	return s.peers.LeaveAll()
}

// SetNSFactory ...
func (s *Server) SetNSFactory(factory func() (rpcutil.Client, error)) {
	s.nsFactory = factory
}

// GetServices ...
func (s *Server) GetServices(name string) (rpcutil.Clients, error) {
	if s.nsFactory == nil {
		return nil, fmt.Errorf("unknown naming-server")
	}
	ns, err := s.nsFactory()
	if err != nil {
		return nil, fmt.Errorf("ns connect failed: %s", err)
	}
	defer ns.Close()
	var addrs []string
	if err := ns.Call("Query", name, &addrs); err != nil {
		return nil, fmt.Errorf("ns query failed: %s", err)
	}
	var res rpcutil.Clients
	for _, addr := range addrs {
		c, err := s.Get(addr, name)
		if err == nil {
			res = append(res, c)
		}
	}
	return res, nil
}

// === for services.Service interface ===

// Name ...
func (s *Server) Name() string {
	return "NodeService"
}

// Rcvr ...
func (s *Server) Rcvr() interface{} {
	return &NodeService{parent: s}
}

// === for HTTP server ===

func (s *Server) handle(ws *websocket.Conn) {
	s.Server.ServeCodec(jsonrpc.NewServerCodec(ws))
}

// ServeHTTP ...
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.httpHandler.ServeHTTP(w, req)
}

// === for Hertbeat ===

func (s *Server) activate() error {
	svcs, err := s.GetServices("NamingService")
	if err != nil {
		return fmt.Errorf("naming service connect failed: %s", err)
	}
	defer svcs.Close()
	for _, svc := range svcs {
		if err := svc.Call("Register", naming.RegisterInfo{
			Addr:     s.SelfAddr(),
			Provides: s.Provides(),
		}, &struct{}{}); err != nil {
			return err
		}
	}
	return nil
}

// Activate ...
func (s *Server) Activate() {
	s.act <- struct{}{}
}
