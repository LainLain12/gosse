package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// IsBanned checks if a user id is in the ban table
func IsBanned(db *sql.DB, id string) bool {
	var exists string
	err := db.QueryRow("SELECT id FROM ban WHERE id=?", id).Scan(&exists)
	return err == nil
}

// Ban represents a banned user by id
type Ban struct {
	ID string `json:"id"`
}

// InitBanTable creates the ban table if it does not exist
func InitBanTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS ban (id TEXT PRIMARY KEY);`
	_, err := db.Exec(query)
	return err
}

// BanHandler handles GET /ban?id=... to ban/check a user
func BanHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}
		var exists string
		err := db.QueryRow("SELECT id FROM ban WHERE id=?", id).Scan(&exists)
		if err == nil { // already banned; ensure past messages removed
			removed := RemoveMessagesByID(id)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":           "already ban",
				"id":               id,
				"removed_messages": removed,
			})
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Not found, insert
		_, err = db.Exec("INSERT INTO ban (id) VALUES (?)", id)
		if err != nil {
			http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		removed := RemoveMessagesByID(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":           "banned",
			"id":               id,
			"removed_messages": removed,
		})
	}
}
