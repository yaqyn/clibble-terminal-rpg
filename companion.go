package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yaqyn/clibble/internal/brand"
	"github.com/yaqyn/clibble/internal/pokeapi"
	"github.com/yaqyn/clibble/internal/webui"
)

func companionHandler(path, authority string) http.Handler {
	_, port, _ := net.SplitHostPort(authority)
	allowed := map[string]bool{"127.0.0.1:" + port: true, "localhost:" + port: true}
	static := http.FileServer(http.FS(webui.Assets()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; font-src 'self'; img-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		w.Header().Set("Cache-Control", "no-store")
		if !allowed[strings.ToLower(r.Host)] {
			http.Error(w, "This collection is available only on the local address.", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Scheme != "http" || !allowed[strings.ToLower(parsed.Host)] {
				http.Error(w, "Cross-origin access is not allowed.", http.StatusForbidden)
				return
			}
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Read-only interface.", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/api/collection" {
			pokemon, err := loadPokedex(path)
			if err != nil {
				http.Error(w, "The collection could not be read. Check the save file with the CLI, then refresh.", http.StatusInternalServerError)
				return
			}
			items := make([]pokeapi.Pokemon, 0, len(pokemon))
			for _, name := range sortedKeys(pokemon) {
				items = append(items, pokemon[name])
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if r.Method == http.MethodHead {
				return
			}
			_ = json.NewEncoder(w).Encode(struct {
				Name    string            `json:"name"`
				Pokemon []pokeapi.Pokemon `json:"pokemon"`
			}{brand.Get().Name, items})
			return
		}
		if r.URL.Path == "/" { // FileServer redirects index.html to the directory; serve it explicitly.
			data, err := http.FS(webui.Assets()).Open("index.html")
			if err != nil {
				http.Error(w, "Interface unavailable.", 500)
				return
			}
			defer data.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if r.Method == http.MethodGet {
				_, _ = io.Copy(w, data)
			}
			return
		}
		static.ServeHTTP(w, r)
	})
}
func serveCompanion(ctx context.Context, path string, port int, out io.Writer) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("-port must be between 0 and 65535")
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("open local companion: %w", err)
	}
	server := &http.Server{Handler: companionHandler(path, listener.Addr().String()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8 << 10}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	fmt.Fprintf(out, "Clibble companion\nOpen http://%s\nRead-only collection view. Catch Pokemon in the CLI, then refresh here.\nPress Ctrl+C to stop.\n", listener.Addr())
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
		return nil
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
