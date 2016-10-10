package websvc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"../../cfg"
)

func logResponse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := httptest.NewRecorder()
		next(c, r)
		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		_, err := w.Write(c.Body.Bytes())
		if err != nil {
			log.ERR("failed to send response: %s", err)
		}

		if cfg.Debug {
			dump, _ := httputil.DumpRequest(r, true)
			headers := ""
			for k, vs := range c.HeaderMap {
				for _, v := range vs {
					headers += fmt.Sprintf("%s: %s\n", k, v)
				}
			}
			log.DEBUG("\n\n%s\n\n%s %d %s\n%s\n%s\n", string(dump),
				r.Proto, c.Code, http.StatusText(c.Code), headers, c.Body.String())
		}

		log.NOTICE("%d -> %-15s %s %s", c.Code, remoteIP(r), r.Method, r.URL.RequestURI())
	}
}
