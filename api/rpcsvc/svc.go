// Package rpcsvc provides handlers for JSON-RPC 2.0.
package rpcsvc

import (
	"net/http"
	"net/rpc"

	"github.com/powerman/narada-go/narada"

	"../../cfg"

	"github.com/powerman/rpc-codec/jsonrpc2"
)

var log = narada.NewLog("rpcsvc: ")

func init() {
	http.Handle(cfg.HTTP.BasePath+"/rpc", jsonrpc2.HTTPHandler(nil))
	if err := rpc.Register(&API{}); err != nil {
		log.Fatal(err)
	}
}

type API struct{}

func (*API) Version(args *struct{}, res *string) error {
	log.DEBUG("RPC: VERSION")
	*res, _ = narada.Version()
	return nil
}
