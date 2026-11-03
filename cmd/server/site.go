package main

import (
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func isAPIHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	want := strings.ToLower(env("LOREMETRY_API_HOST", "api.loremetry.com"))
	if host == want {
		return true
	}
	return strings.HasPrefix(host, "api.")
}

func websiteDir() string {
	return env("WEBSITE_DIR", filepath.Join("website", "dist"))
}

func websiteHandler(dir string) http.Handler {
	index := filepath.Join(dir, "index.html")
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if rel != "." && rel != "" {
			full := filepath.Join(dir, filepath.FromSlash(rel))
			if strings.HasPrefix(full, dir+string(os.PathSeparator)) || full == dir {
				if st, err := os.Stat(full); err == nil && !st.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}
		http.ServeFile(w, r, index)
	})
}

func routeByHost(api, site http.Handler, health http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			health(w, r)
			return
		}
		if isAPIHost(r.Host) || site == nil {
			api.ServeHTTP(w, r)
			return
		}
		site.ServeHTTP(w, r)
	})
}
