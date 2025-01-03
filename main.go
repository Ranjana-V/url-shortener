package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// In-memory storage for URLs
var urlStore = make(map[string]string)
var mu sync.Mutex

func main() {
	// Serve the index.html file
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/shorten", ShortenURLHandler)
	http.HandleFunc("/redirect/", RedirectHandler)

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HomeHandler: Displays the HTML form for shortening URLs
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// ShortenURLHandler: Creates a short URL
func ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		OriginalURL string `json:"original_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode(request.OriginalURL)

	mu.Lock()
	urlStore[shortCode] = request.OriginalURL
	mu.Unlock()

	response := map[string]string{
		"short_url": "http://localhost:8080/redirect/" + shortCode,
	}
	json.NewEncoder(w).Encode(response)
}

// RedirectHandler: Redirects to the original URL
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[len("/redirect/"):]

	mu.Lock()
	originalURL, exists := urlStore[shortCode]
	mu.Unlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// generateShortCode: Simple hash function for short codes
func generateShortCode(url string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(url)))[:6]
}
