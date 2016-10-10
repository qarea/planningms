// Package websvc provides handlers for Web.
package websvc

import (
	"fmt"
	"net/http"

	"github.com/powerman/narada-go/narada"

	"../../cfg"
)

var log = narada.NewLog("")

func init() {
	http.Handle(cfg.HTTP.BasePath+"/web", http.StripPrefix(cfg.HTTP.BasePath, http.HandlerFunc(logResponse(web))))
}

func web(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}
