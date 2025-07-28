package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ChatSSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// On first connect, send all 50 messages as SSE events (data: ...)
	chatMu.Lock()
	all := make([]any, len(chatMessages))
	copy(all, chatMessages)
	chatMu.Unlock()
	for _, msg := range all {
		b, err := json.Marshal(msg)
		if err == nil {
			fmt.Fprintf(w, "data: %s\n\n", string(b))
		}
	}
	flusher.Flush()

	// Track the last sent message index
	lastIdx := len(all)

	// Broadcast new messages as they arrive (as single JSON object, not array)
	notify := r.Context().Done()
	msgTicker := time.NewTicker(1 * time.Second)
	pingTicker := time.NewTicker(15 * time.Second)
	defer msgTicker.Stop()
	defer pingTicker.Stop()
	for {
		select {
		case <-notify:
			return
		case <-msgTicker.C:
			chatMu.Lock()
			if lastIdx < len(chatMessages) {
				for _, msg := range chatMessages[lastIdx:] {
					b, err := json.Marshal(msg)
					if err == nil {
						fmt.Fprintf(w, "data: %s\n\n", string(b))
					}
				}
				flusher.Flush()
				lastIdx = len(chatMessages)
			}
			chatMu.Unlock()
		case <-pingTicker.C:
			// Send SSE comment as keepalive (ping)
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// // ChatSSEHandler streams chat messages via SSE
// func ChatSSEHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")

// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
// 		return
// 	}

// 	// On first connect, send all 50 messages as a JSON array
// 	chatMu.Lock()
// 	all := make([]any, len(chatMessages))
// 	copy(all, chatMessages)
// 	chatMu.Unlock()
// 	b, err := json.Marshal(all)
// 	if err == nil {
// 		fmt.Fprintf(w, "%s\n", string(b))
// 		flusher.Flush()
// 	}

// 	// Track the last sent message index
// 	lastIdx := len(all)

// 	// Broadcast new messages as they arrive (as single JSON object, not array)
// 	notify := r.Context().Done()
// 	msgTicker := time.NewTicker(1 * time.Second)
// 	pingTicker := time.NewTicker(15 * time.Second)
// 	defer msgTicker.Stop()
// 	defer pingTicker.Stop()
// 	for {
// 		select {
// 		case <-notify:
// 			return
// 		case <-msgTicker.C:
// 			chatMu.Lock()
// 			if lastIdx < len(chatMessages) {
// 				for _, msg := range chatMessages[lastIdx:] {
// 					b, err := json.Marshal(msg)
// 					if err == nil {
// 						fmt.Fprintf(w, "%s\n", string(b))
// 					}
// 				}
// 				flusher.Flush()
// 				lastIdx = len(chatMessages)
// 			}
// 			chatMu.Unlock()
// 		case <-pingTicker.C:
// 			// Send SSE comment as keepalive (ping)
// 			fmt.Fprintf(w, ": ping\n\n")
// 			flusher.Flush()
// 		}
// 	}
// }
