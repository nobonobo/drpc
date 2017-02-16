package node

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"

	"golang.org/x/net/websocket"
)

var (
	// DefaultMaxConns ノードごとの最大接続数
	DefaultMaxConns = 2
	// DefaultURLPath ...
	DefaultURLPath = "/ws"
)

// DefaultLocalFactory ...
var DefaultLocalFactory = func(n *Server) (*rpc.Client, error) {
	cl, cr := net.Pipe()
	go n.Server.ServeCodec(jsonrpc.NewServerCodec(cl))
	return jsonrpc.NewClient(cr), nil
}

// DefaultRemoteFactory ...
var DefaultRemoteFactory = func(addr string) (*rpc.Client, error) {
	u := url.URL{}
	u.Scheme = "ws"
	u.Host = addr
	u.Path = DefaultURLPath
	uri := u.String()
	u.Scheme = "http"
	u.Path = "/"
	origin := u.String()
	conn, err := websocket.Dial(uri, "", origin)
	if err != nil {
		return nil, err
	}
	return jsonrpc.NewClient(conn), nil
}
