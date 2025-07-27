package lottosociety

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// GetLottoHandler handles GET /getlotto?date=... or ?last=true to return lotto rows by date, all, or just the latest
func GetLottoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		last := r.URL.Query().Get("last")
		var rows *sql.Rows
		var err error
		if last == "true" {
			rows, err = db.Query("SELECT date, thaidate, fnum, snum, id, text FROM lottosociety ORDER BY date DESC LIMIT 1")
		} else if date != "" {
			rows, err = db.Query("SELECT date, thaidate, fnum, snum, id, text FROM lottosociety WHERE date=?", date)
		} else {
			rows, err = db.Query("SELECT date, thaidate, fnum, snum, id, text FROM lottosociety ORDER BY date DESC")
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var all []LottoSociety
		for rows.Next() {
			var l LottoSociety
			err := rows.Scan(&l.Date, &l.ThaiDate, &l.FNum, &l.SNum, &l.ID, &l.Text)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			all = append(all, l)
		}
		w.Header().Set("Content-Type", "application/json")
		if last == "true" && len(all) > 0 {
			json.NewEncoder(w).Encode(all[0])
		} else {
			json.NewEncoder(w).Encode(all)
		}
	}
}
