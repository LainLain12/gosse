package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// SendMessageHandler returns a handler that stores a message if user not banned
func SendMessageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var msg map[string]any
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Extract user id from several possible keys or query param
		extractID := func() string {
			keys := []string{"id", "userId", "userid", "user_id"}
			for _, k := range keys {
				if v, ok := msg[k]; ok {
					if s, ok := v.(string); ok {
						return strings.TrimSpace(s)
					}
				}
			}
			qp := strings.TrimSpace(r.URL.Query().Get("id"))
			return qp
		}
		id := extractID()
		if id == "" {
			http.Error(w, "Missing id in message", http.StatusBadRequest)
			return
		}
		if IsBanned(db, id) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"status":  "banned",
				"message": "you are banned",
			})
			return
		}
		// Attach normalized id back into message to ensure consistency
		msg["id"] = id
		// First store (with dedup) then publish only if actually stored
		if added := AddChatMessage(msg); added {
			publish(msg)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"message": msg,
		})
	}
}
