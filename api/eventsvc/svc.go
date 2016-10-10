// Package eventsvc provides handlers for EventSource.
package eventsvc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/powerman/narada-go/narada"

	"../../cfg"
)

var log = narada.NewLog("")

func init() {
	http.HandleFunc(cfg.HTTP.BasePath+"/events", events)
}

func events(w http.ResponseWriter, r *http.Request) {
	log.DEBUG("EV: connected")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Expires", "Sat, 01 Jan 2000 00:00:00 GMT")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()
	for {
		time.Sleep(1 * time.Second)
		_, err := fmt.Fprintf(w, "data: Now %s\n\n", time.Now())
		if err != nil {
			log.DEBUG("EV: disconnected")
			return
		}
		w.(http.Flusher).Flush()
	}
}
