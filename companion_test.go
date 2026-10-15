package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaqyn/dexrove/internal/pokeapi"
)

func TestCompanionReadsCollectionAndServesEmbeddedAssets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey", Height: 3, Weight: 18}})
	handler := companionHandler(path, "127.0.0.1:8787")
	for _, route := range []string{"/", "/app.js", "/styles.css", "/tokens.css", "/mark.svg", "/manrope.ttf", "/api/collection"} {
		req := httptest.NewRequest("GET", "http://127.0.0.1:8787"+route, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d", route, response.Code)
		}
		if response.Header().Get("Content-Security-Policy") == "" {
			t.Fatal("missing CSP")
		}
		if route == "/api/collection" {
			var body struct{ Pokemon []pokeapi.Pokemon }
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if len(body.Pokemon) != 1 || body.Pokemon[0].Name != "pidgey" {
				t.Fatal(response.Body.String())
			}
			if strings.Contains(response.Body.String(), path) {
				t.Fatal("save path exposed")
			}
		}
	}
}
func TestCompanionRejectsForeignHostsOriginsAndWrites(t *testing.T) {
	handler := companionHandler(filepath.Join(t.TempDir(), "collection.json"), "127.0.0.1:8787")
	for _, tc := range []struct {
		host, origin, method string
		code                 int
	}{
		{"evil.example:8787", "", "GET", 403},
		{"127.0.0.1:8787", "https://evil.example", "GET", 403},
		{"127.0.0.1:8787", "null", "GET", 403},
		{"localhost:8787", "", "POST", 405},
		{"localhost:8787", "", "DELETE", 405},
		{"localhost:8787", "http://localhost:8787", "GET", 200},
	} {
		req := httptest.NewRequest(tc.method, "http://"+tc.host+"/api/collection", nil)
		req.Header.Set("Origin", tc.origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != tc.code {
			t.Fatalf("%+v: %d", tc, response.Code)
		}
	}
}
func TestCompanionReportsCorruptionWithoutLeakingOrChangingSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-collection.json")
	os.WriteFile(path, []byte("{broken"), 0600)
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection", nil))
	if response.Code != 500 || strings.Contains(response.Body.String(), path) {
		t.Fatal(response.Body.String())
	}
	data, _ := os.ReadFile(path)
	if string(data) != "{broken" {
		t.Fatal("read-only endpoint changed save")
	}
}
