package futurepaper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// UploadPaperImageHandler handles POST /futurepaper/uploadimage to upload an image to futurepaper/images/
func UploadPaperImageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			http.Error(w, "Internal server error (panic)", http.StatusInternalServerError)
		}
	}()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create images folder if not exists
	if err := os.MkdirAll("futurepaper/images", 0755); err != nil {
		http.Error(w, "Failed to create images dir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read all image data from body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read image: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Use extension from Content-Type if possible, else default to .png
	ext := ".png"
	switch r.Header.Get("Content-Type") {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	}

	fname := fmt.Sprintf("%f%s", float64(time.Now().UnixNano())/1e9, ext)
	fpath := "futurepaper/images/" + fname
	out, err := os.Create(fpath)
	if err != nil {
		http.Error(w, "Failed to save image: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = out.Write(body)
	if err != nil {
		http.Error(w, "Failed to write image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	relUrl := "/futurepaper/images/" + fname
	fullUrl := scheme + "://" + host + relUrl

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"imagename": fname,
		"url":       fullUrl,
	})
}
