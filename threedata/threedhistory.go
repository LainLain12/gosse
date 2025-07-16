package threedata

import (
	"database/sql"
	"encoding/json"

	"net/http"
)

// ThreedDataHandler handles GET /threeddata and returns all rows as JSON
func ThreedDataHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT date, result FROM threeddata`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var all []ThreedData
		for rows.Next() {
			var d ThreedData
			err := rows.Scan(&d.Date, &d.Result)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			all = append(all, d)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}
