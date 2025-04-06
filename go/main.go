package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// validTokens holds the allowed tokens as a set.
var validTokens map[string]struct{}

// loadValidTokens reads the VALID_TOKENS environment variable and populates validTokens.
func loadValidTokens() {
	validTokens = make(map[string]struct{})
	tokens := os.Getenv("VALID_TOKENS")
	for _, token := range strings.Split(tokens, ",") {
		token = strings.TrimSpace(token)
		if token != "" {
			validTokens[token] = struct{}{}
		}
	}
}

func main() {
	// Load tokens from the environment
	loadValidTokens()

	// Handle the protected page
	http.HandleFunc("/protected", protectedHandler)

	// Serve static files from the "site" directory for the unprotected Jekyll site.
	fs := http.FileServer(http.Dir("site"))
	http.Handle("/", fs)

	// Use port 5000 or PORT env variable if provided
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Printf("Server listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// protectedHandler checks for a valid token and serves the protected page.
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if _, ok := validTokens[token]; !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, "protected/secret.html")
}
