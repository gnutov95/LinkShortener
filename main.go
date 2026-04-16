package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	storage = make(map[string]string)
	mu      sync.Mutex
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"short_url"`
}

func keyGen() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, 6)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

func main() {

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Request

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}

		url := req.URL

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		key := "/" + keyGen()

		mu.Lock()
		storage[key] = url
		mu.Unlock()

		short := "http://localhost:2222" + key

		json.NewEncoder(w).Encode(Response{
			ShortURL: short,
		})
	})

	// redirect handler (ВАЖНО: catch-all)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// если это API или статика — пропускаем
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}

		mu.Lock()
		dest, ok := storage[r.URL.Path]
		mu.Unlock()

		if ok {
			http.Redirect(w, r, dest, http.StatusMovedPermanently)
			return
		}

		http.NotFound(w, r)
	})

	fmt.Println("http://localhost:2222")

	http.ListenAndServe(":2222", mux)
}
