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
			// Update row by date
			res, err := db.Exec("UPDATE lottosociety SET thaidate=?, fnum=?, snum=?, id=?, text=? WHERE date=?", req.ThaiDate, req.FNum, req.SNum, req.ID, req.Text, req.Date)
			if err != nil {
				http.Error(w, "Database update error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				// If no row updated, insert new
				_, err := db.Exec("INSERT INTO lottosociety (date, thaidate, fnum, snum, id, text) VALUES (?, ?, ?, ?, ?, ?)", req.Date, req.ThaiDate, req.FNum, req.SNum, req.ID, req.Text)
				if err != nil {
					http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "updated or inserted",
				"date":   req.Date,
			})
			return
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
