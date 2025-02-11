package main

import (
	// 	"net/http"
	// 	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	t.Name()
}

// func TestHTTPSRedirectMiddleware(t *testing.T) {
// 	handler := httpsRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}))

// 	req := httptest.NewRequest("GET", "http://localhost/", nil)
// 	w := httptest.NewRecorder()

// 	handler.ServeHTTP(w, req)

// 	res := w.Result()
// 	if res.StatusCode != http.StatusMovedPermanently {
// 		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, res.StatusCode)
// 	}

// 	expectedLocation := "https://localhost/"
// 	if location := res.Header.Get("Location"); location != expectedLocation {
// 		t.Errorf("expected Location header %q, got %q", expectedLocation, location)
// 	}
// }

// func TestRootHandler(t *testing.T) {
// 	req := httptest.NewRequest("GET", "https://localhost/", nil)
// 	w := httptest.NewRecorder()

// 	rootHandler(w, req)

// 	res := w.Result()
// 	if res.StatusCode != http.StatusOK {
// 		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
// 	}
// }
