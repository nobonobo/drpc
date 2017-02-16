package node

import (
	"fmt"

	"github.com/nobonobo/drpc/rpcutil"
	"github.com/nobonobo/drpc/services/naming"
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
	svcs, err := ns.parent.GetServices("NamingService")
	if err != nil {
		return fmt.Errorf("naming service connect failed: %s", err)
	}
	defer svcs.Close()
	for _, svc := range svcs {
		if err := svc.Call("Register", naming.RegisterInfo{
			Addr:     ns.parent.SelfAddr(),
			Provides: ns.parent.Provides(),
		}, &struct{}{}); err != nil {
			return err
		}
	}
	return nil
}

// Bye ...
func (ns *NodeService) Bye(addr string, none *struct{}) error {
	return ns.parent.Leave(addr)
}
