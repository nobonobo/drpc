package node

import (
	"github.com/nobonobo/drpc/rpcutil"
)

// NodeService Remote Stub
type NodeService struct {
	parent *Server
}

// Invite ...
func (ns *NodeService) Invite(nsAddr string, none *struct{}) error {
	if err := ns.parent.Join(nsAddr); err != nil {
		return err
	}
	ns.parent.SetNSFactory(func() (rpcutil.Client, error) {
		return ns.parent.Get(nsAddr, "NamingService")
	})
	ns.parent.Activate()
	return nil
}

// Bye ...
func (ns *NodeService) Bye(addr string, none *struct{}) error {
	return ns.parent.Leave(addr)
}
