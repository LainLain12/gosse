package Live

import (
	"encoding/json"
	"gosse/twoddata"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// AddLiveDataHandler handles POST /addLiveData and stores the data in memory
func AddLiveDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	var data Live
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	liveDataMu.Lock()
	liveDataStore = []Live{data}
	jdata, err := json.Marshal(liveDataStore)
	if err != nil {
		return
	}
	os.WriteFile("live.json", jdata, 0644)
	liveDataMu.Unlock()

	// --- DB insert logic ---
	// 1. If todaydate data not in DB
	// 2. If time > 16:30
	// 3. If post data.Live does not contain "-"
	now := time.Now()
	if now.Hour() > 16 || (now.Hour() == 16 && now.Minute() >= 30) {
		if !strings.Contains(data.Live, "-") {
			// Open DB (assume twoddata.db in root, or adjust as needed)
			db := twoddata.InitDB("twoddata.db")
			defer db.Close()
			// Check if today's data exists
			var count int
			dateStr := data.Date
			err := db.QueryRow("SELECT COUNT(*) FROM twoddata WHERE date = ?", dateStr).Scan(&count)
			if err == nil && count == 0 {
				if data.Eresult == "--" {
					return
				}
				// Insert new row
				_, err := db.Exec(`INSERT INTO twoddata (mset, mvalue, mresult, eset, evalue, eresult, tmodern, tinernet, nmodern, ninternet, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					data.Mset, data.Mvalue, data.Mresult, data.Eset, data.Evalue, data.Eresult, data.Tmodern, data.Tinternet, data.Nmodern, data.Ninternet, dateStr)
				if err != nil {
					// Log or handle error as needed
				}
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("ok"))
}
