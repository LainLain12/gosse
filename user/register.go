package user

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// RegisterUserHandler handles POST /register to add a new useraccount if id not exists
func RegisterUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get id from query parameter
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		// Check if user already exists
		var exists string
		err := db.QueryRow("SELECT id FROM useraccount WHERE id=?", id).Scan(&exists)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "already registered",
				"id":     id,
			})
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse JSON body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body: "+err.Error(), http.StatusBadRequest)
			return
		}
		var user UserAccount
		if err := json.Unmarshal(body, &user); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Insert new user
		_, err = db.Exec("INSERT INTO useraccount (id, name, profile_pic, email) VALUES (?, ?, ?, ?)", user.ID, user.Name, user.ProfilePic, user.Email)
		if err != nil {
			http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "registered",
			"id":     user.ID,
		})
	}
}
