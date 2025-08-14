package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"
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
		id, ok := msg["id"].(string)
		if !ok || id == "" {
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
		publish(msg)
		AddChatMessage(msg)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"message": msg,
		})
	}
}
