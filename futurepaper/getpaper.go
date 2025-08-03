package futurepaper

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetPaperHandler returns a JSON list of image URLs in futurepaper/images/
func GetPaperHandler(w http.ResponseWriter, r *http.Request) {
	imageDir := "futurepaper/images"
	exts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".webp": true}
	var files []string
	entries, err := os.ReadDir(imageDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	baseURL := scheme + "://" + host + "/futurepaper/images/"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if exts[ext] {
			files = append(files, baseURL+entry.Name())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// GetAllPaperHandler returns JSON with image names in daily/ and weekly/ folders
func GetAllPaperHandler(w http.ResponseWriter, r *http.Request) {
	dirs := []string{"daily", "weekly"}
	exts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".webp": true}

	result := map[string][]string{"daily": {}, "weekly": {}}
	for _, dir := range dirs {
		fullDir := filepath.Join("futurepaper/images", dir)
		entries, err := os.ReadDir(fullDir)
		if err != nil {
			result[dir] = []string{}
			continue
		}
		var names []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if exts[ext] {
				names = append(names, entry.Name())
			}
		}
		result[dir] = names
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
