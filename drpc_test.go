package drpc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nobonobo/drpc/rpcutil"
	"github.com/nobonobo/drpc/services/naming"
	"github.com/nobonobo/drpc/services/node"
)

type testNode struct {
	*node.Server
	ts *httptest.Server
}
type testNodes []*testNode

func (n testNodes) Close() {
	for _, one := range n {
		one.Server.Close()
		one.ts.Close()
	}
}

func newTestNodes(num int) (testNodes, error) {
	nodes := testNodes{}
	for i := 0; i < num; i++ {
		mux := http.NewServeMux()
		ts := httptest.NewServer(mux)
		n, err := node.New(ts.Listener.Addr().String())
		if err != nil {
			nodes.Close()
			return nil, err
		}
		mux.Handle(node.DefaultURLPath, n)
		nodes = append(nodes, &testNode{Server: n, ts: ts})
	}
	return nodes, nil
}

func TestNormal(t *testing.T) {
	nodes, err := newTestNodes(3)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer nodes.Close()
	master, slaves := nodes[0], nodes[1:]
	ns := naming.New()
	t.Log(master.Register(ns))
	ns.Update(naming.RegisterInfo{
		Addr:     master.SelfAddr(),
		Provides: master.Provides(),
	})
	master.SetNSFactory(func() (rpcutil.Client, error) {
		return master.Get(master.SelfAddr(), "NamingService")
	})
	for _, node := range nodes {
		t.Log(node.SelfAddr(), node.Provides())
	}
	for _, node := range slaves {
		master.Join(node.SelfAddr())
		c, err := master.Get(node.SelfAddr(), "NodeService")
		if err != nil {
			t.Log(err)
			continue
		}
		defer c.Close()
		t.Log(c.Call("Invite", master.SelfAddr(), &struct{}{}))
	}
	svcs, err := master.GetServices("NodeService")
	if err != nil {
		t.Log(err)
	} else {
		t.Log(svcs)
		svcs.Close()
	}
	svcs, err = master.GetServices("NamingService")
	if err != nil {
		t.Log(err)
	} else {
		t.Log(svcs)
		svcs.Close()
	}
}
