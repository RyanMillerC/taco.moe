package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// statusResponseWriter wraps http.ResponseWriter to capture the status code.
type statusResponseWriter struct {
        http.ResponseWriter
        statusCode int
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *statusResponseWriter) WriteHeader(code int) {
        w.statusCode = code
        w.ResponseWriter.WriteHeader(code)
}

// Log http requests to stdout
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to capture the status code.
		sw := &statusResponseWriter{ResponseWriter: w, statusCode: 200} // default to 200 if WriteHeader is not called
		
		// Process the request.
		next.ServeHTTP(sw, r)
		
		// Log the details.
		log.Printf(
			"%s \"%s %s\" %d \"%s\" \"%s\"",
			getIP(r),                   // Remote IP (with proxy support)
			r.Method,                   // HTTP Method
			r.URL.RequestURI(),         // Full URL with query params
			sw.statusCode,              // HTTP Status Code
			r.Header.Get("User-Agent"), // User Agent
			r.Referer(),                // Referrer
		)
	})
}

// Get http.Request IP from X-Forwarded-For header if it exists. Otherwise,
// fall back to RemoteAddr
func getIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	return r.RemoteAddr
}

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

// Checks for a valid token and serves the protected page.
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if _, ok := validTokens[token]; !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, "personal-user-manual/index.html")
}

func main() {
	// Load tokens from the environment
	loadValidTokens()

	// Handle the protected page
	http.HandleFunc("/pum", protectedHandler)

	// Serve static files from the "site" directory for the unprotected Jekyll site.
	fs := http.FileServer(http.Dir("site"))
	http.Handle("/", fs)

	// Use port 5000 or PORT env variable if provided
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Printf("Server listening on port %s", port)
	err := http.ListenAndServe(":"+port, loggingHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}
