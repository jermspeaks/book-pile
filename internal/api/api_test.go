package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jermspeaks/book-pile/internal/store"
)

func TestOwnershipAndPiles(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	h := New(db, dataDir, "")

	book := postJSON(t, h, "/api/books", map[string]any{
		"title":   "A Wizard of Earthsea",
		"authors": []string{"Ursula K. Le Guin"},
	})
	if book["ownership"] != "coveting" {
		t.Fatalf("expected coveting, got %v", book["ownership"])
	}

	patched := patchJSON(t, h, "/api/books/"+idStr(book), map[string]any{
		"ownedPhysical": true,
		"wantRating":    4,
		"readStatus":    "unread",
	})
	if patched["ownership"] != "physical" {
		t.Fatalf("expected physical, got %v", patched["ownership"])
	}
	if patched["wantRating"].(float64) != 4 {
		t.Fatalf("want=%v", patched["wantRating"])
	}

	pile := postJSON(t, h, "/api/piles", map[string]any{"name": "Bedside"})
	putJSON(t, h, "/api/books/"+idStr(book)+"/piles", map[string]any{
		"pileIds": []any{pile["id"]},
	})
	got := getJSON(t, h, "/api/books/"+idStr(book))
	piles, _ := got["piles"].([]any)
	if len(piles) != 1 {
		t.Fatalf("piles=%v", got["piles"])
	}
	col := postJSON(t, h, "/api/collections", map[string]any{"name": "Reread someday"})
	putJSON(t, h, "/api/books/"+idStr(book)+"/collections", map[string]any{
		"collectionIds": []any{col["id"]},
	})
	got = getJSON(t, h, "/api/books/"+idStr(book))
	cols, _ := got["collections"].([]any)
	if len(cols) != 1 {
		t.Fatalf("collections=%v", got["collections"])
	}

	list := getJSONList(t, h, "/api/books?ownership=physical")
	if len(list) != 1 {
		t.Fatalf("filtered list=%d", len(list))
	}
	empty := getJSONList(t, h, "/api/collections")
	if empty == nil {
		t.Fatal("collections should be [] not null")
	}
}

func postJSON(t *testing.T, h http.Handler, path string, body any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		t.Fatalf("POST %s: %d %s", path, rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func patchJSON(t *testing.T, h http.Handler, path string, body any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		t.Fatalf("PATCH %s: %d %s", path, rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putJSON(t *testing.T, h http.Handler, path string, body any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		t.Fatalf("PUT %s: %d %s", path, rec.Code, rec.Body.String())
	}
}

func getJSON(t *testing.T, h http.Handler, path string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		t.Fatalf("GET %s: %d %s", path, rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func getJSONList(t *testing.T, h http.Handler, path string) []any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		t.Fatalf("GET %s: %d %s", path, rec.Code, rec.Body.String())
	}
	var out []any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func idStr(m map[string]any) string {
	switch v := m["id"].(type) {
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return "0"
	}
}
