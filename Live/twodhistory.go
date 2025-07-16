package Live

import (
	"database/sql"
	"encoding/json"
	"gosse/twoddata"
	"net/http"
)

// TwoddataHandler handles GET /twoddata and returns all rows as JSON
func TwoddataHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, mset, mvalue, mresult, eset, evalue, eresult, tmodern, tinernet, nmodern, ninternet, date FROM twoddata`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var all []twoddata.TwodData
		for rows.Next() {
			var d twoddata.TwodData
			err := rows.Scan(&d.ID, &d.MSet, &d.MValue, &d.MResult, &d.ESet, &d.EValue, &d.EResult, &d.TModern, &d.TInernet, &d.NModern, &d.NInernet, &d.Date)
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
