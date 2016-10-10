package websvc

import (
	_ "log"
	"net"
	"net/http"

	"../../cfg"
)

func remoteIP(r *http.Request) (ip string) {
	if cfg.HTTP.RealIPHeader == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	} else {
		ip = r.Header.Get(cfg.HTTP.RealIPHeader)
	}
	if ip == "" {
		log.Fatal("failed to detect remote IP, check config/real_ip_header")
	}
	return
}
