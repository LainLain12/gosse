package Live

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type api struct {
	Data        interface{} `json:"data"`
	ClientCount int         `json:"clinetcount"`
}

var Clinets []string
var clientsMu sync.Mutex

// Global slice to hold all connected SSE clients
// In-memory live data store and mutex

// LiveDataSSEHandler streams the current liveDataStore as JSON every second
func LiveDataSSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	clientsMu.Lock()
	Clinets = append(Clinets, r.RemoteAddr)
	clientsMu.Unlock()

	var prevLive string
	for {
		select {
		case <-time.After(1 * time.Second):
			liveDataMu.Lock()
			var currLive string
			if len(liveDataStore) > 0 {
				currLive = liveDataStore[0].Live
			}
			if currLive != prevLive {
				// Convert to JSON
				var ap api
				ap.Data = liveDataStore
				ap.ClientCount = len(Clinets)
				data, _ := json.Marshal(ap)
				if _, err := w.Write([]byte("data: ")); err != nil {
					liveDataMu.Unlock()
					return
				}
				if _, err := w.Write(data); err != nil {
					liveDataMu.Unlock()
					return
				}
				if _, err := w.Write([]byte("\n\n")); err != nil {
					liveDataMu.Unlock()
					return
				}
				flusher.Flush()
				prevLive = currLive
			}
			liveDataMu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// LiveDataPageHandler serves an HTML page that shows the live JSON data
func LiveDataPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html><body>
<pre id="json"></pre>
<script>
var es = new EventSource('/livedata/sse');
es.onmessage = function(e) {
  document.getElementById('json').textContent = e.data;
};
</script>
</body></html>`))
}
