package lottosociety

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// LottoSociety represents a lottery entry
// LottoSociety represents the structure of the lottery data

// AddOrUpdateLottoHandler handles POST /addlotto to update by date or insert new row
func AddOrUpdateLottoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req LottoSociety
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.Date != "" {
			if req.Date == "Invalid Date" {
				http.Error(w, "Invalid date value", http.StatusBadRequest)
				return
			}
			// Check if date exists
			var exists string
			err := db.QueryRow("SELECT date FROM lottosociety WHERE date=?", req.Date).Scan(&exists)
			if err == nil {
				// Date exists, update row
				_, err := db.Exec("UPDATE lottosociety SET thaidate=?, fnum=?, snum=?, id=?, text=? WHERE date=?", req.ThaiDate, req.FNum, req.SNum, req.ID, req.Text, req.Date)
				if err != nil {
					http.Error(w, "Database update error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "updated",
					"date":   req.Date,
				})
				return
			} else if err == sql.ErrNoRows {
				// Date not found, insert new row
				_, err := db.Exec("INSERT INTO lottosociety (date, thaidate, fnum, snum, id, text) VALUES (?, ?, ?, ?, ?, ?)", req.Date, req.ThaiDate, req.FNum, req.SNum, req.ID, req.Text)
				if err != nil {
					http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "inserted",
				})
				return
			} else {
				http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// No date, insert new row
		_, err := db.Exec("INSERT INTO lottosociety (date, thaidate, fnum, snum, id, text) VALUES (?, ?, ?, ?, ?, ?)", req.Date, req.ThaiDate, req.FNum, req.SNum, req.ID, req.Text)
		if err != nil {
			http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "inserted",
		})
	}
}
