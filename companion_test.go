package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
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
		if response.Header().Get("Permissions-Policy") == "" || response.Header().Get("Cross-Origin-Resource-Policy") != "same-origin" {
			t.Fatal("missing browser isolation headers")
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
func TestCompanionCollectionSearch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{
		"pidgey": {Name: "pidgey"}, "pikachu": {Name: "pikachu"},
	})
	handler := companionHandler(path, "127.0.0.1:8787")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?q=PIKA", nil))
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "pidgey") || !strings.Contains(response.Body.String(), "pikachu") {
		t.Fatal(response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?q="+strings.Repeat("x", 101), nil))
	if response.Code != http.StatusBadRequest {
		t.Fatal(response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?q=pika&q=pidgey", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatal("duplicate search accepted", response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?query=pika", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatal("unknown search accepted", response.Code)
	}
}
func TestCompanionCollectionLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{"alpha": {Name: "alpha"}, "beta": {Name: "beta"}})
	handler := companionHandler(path, "127.0.0.1:8787")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?limit=1", nil))
	var body struct{ Pokemon []pokeapi.Pokemon }
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || response.Code != 200 || len(body.Pokemon) != 1 || body.Pokemon[0].Name != "alpha" {
		t.Fatalf("limited response: %d %s", response.Code, response.Body.String())
	}
	for _, value := range []string{"0", "501", "many"} {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?limit="+value, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("accepted limit %q: %d", value, response.Code)
		}
	}
}
func TestCompanionCollectionSortsDescendingBeforeLimiting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{"alpha": {Name: "alpha"}, "beta": {Name: "beta"}})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?sort=name-desc&limit=1", nil))
	if response.Code != 200 || !strings.Contains(response.Body.String(), "beta") || strings.Contains(response.Body.String(), "alpha") {
		t.Fatalf("descending response: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?sort=random", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("accepted invalid sort: %d", response.Code)
	}
}
func TestCompanionCollectionOffset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{"alpha": {Name: "alpha"}, "beta": {Name: "beta"}})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?offset=1&limit=1", nil))
	if response.Code != 200 || strings.Contains(response.Body.String(), "alpha") || !strings.Contains(response.Body.String(), "beta") {
		t.Fatalf("offset response: %d %s", response.Code, response.Body.String())
	}
	for _, value := range []string{"-1", "100001", "many"} {
		response = httptest.NewRecorder()
		companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?offset="+value, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("accepted offset %q: %d", value, response.Code)
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
func TestCompanionCollectionSearchIncludesTypes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	var pidgey pokeapi.Pokemon
	if err := json.Unmarshal([]byte(`{"name":"pidgey","types":[{"type":{"name":"flying"}}]}`), &pidgey); err != nil {
		t.Fatal(err)
	}
	savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": pidgey, "pikachu": {Name: "pikachu"}})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?q=FLYING", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "pidgey") || strings.Contains(response.Body.String(), "pikachu") {
		t.Fatal(response.Code, response.Body.String())
	}
}
func TestCompanionCollectionFiltersByExactType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	var flying, electric pokeapi.Pokemon
	if err := json.Unmarshal([]byte(`{"name":"pidgey","types":[{"type":{"name":"flying"}}]}`), &flying); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"name":"pikachu","types":[{"type":{"name":"electric"}}]}`), &electric); err != nil {
		t.Fatal(err)
	}
	savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": flying, "pikachu": electric})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?type=FLYING", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "pidgey") || strings.Contains(response.Body.String(), "pikachu") {
		t.Fatal(response.Code, response.Body.String())
	}
}
func TestCompanionCollectionFiltersByExactName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}, "pidgeotto": {Name: "pidgeotto"}})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?name=PIDGEY", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "pidgey") || strings.Contains(response.Body.String(), "pidgeotto") {
		t.Fatal(response.Code, response.Body.String())
	}
}
func TestCompanionCollectionSearchCombinesWordsAcrossNameAndType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	var pidgey, pikachu pokeapi.Pokemon
	if err := json.Unmarshal([]byte(`{"name":"pidgey","types":[{"type":{"name":"flying"}}]}`), &pidgey); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"name":"pikachu","types":[{"type":{"name":"electric"}}]}`), &pikachu); err != nil {
		t.Fatal(err)
	}
	savePokedex(path, map[string]pokeapi.Pokemon{"pidgey": pidgey, "pikachu": pikachu})
	response := httptest.NewRecorder()
	companionHandler(path, "127.0.0.1:8787").ServeHTTP(response, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/collection?q=pidge+flying", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "pidgey") || strings.Contains(response.Body.String(), "pikachu") {
		t.Fatal(response.Code, response.Body.String())
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
