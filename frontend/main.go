package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/", fs)

	// Proxy API calls
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		var target *url.URL
		switch r.URL.Path {
		case "/api/user/auth/login", "/api/user/register":
			target, _ = url.Parse("http://user-service:8080")
		case "/api/order":
			target, _ = url.Parse("http://order-service:8080")
		case "/api/portfolio":
			target, _ = url.Parse("http://portfolio-service:8080")
		default:
			http.NotFound(w, r)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	})

	log.Println("Frontend service running on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
