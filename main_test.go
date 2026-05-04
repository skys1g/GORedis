package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setup() {
	dataMap = make(map[string]any)
}

func TestSET(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodPost, "/set?key=name&value=john", nil)
	w := httptest.NewRecorder()
	SET(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if dataMap["name"] != "john" {
		t.Errorf("expected john, got %v", dataMap["name"])
	}
}

func TestSET_WrongMethod(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodGet, "/set?key=name&value=john", nil)
	w := httptest.NewRecorder()
	SET(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestSET_NoKey(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodPost, "/set?value=john", nil)
	w := httptest.NewRecorder()
	SET(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGET(t *testing.T) {
	setup()
	dataMap["name"] = "john"
	req := httptest.NewRequest(http.MethodGet, "/get?key=name", nil)
	w := httptest.NewRecorder()
	GET(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "john" {
		t.Errorf("expected john, got %s", w.Body.String())
	}
}

func TestGET_NotFound(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodGet, "/get?key=missing", nil)
	w := httptest.NewRecorder()
	GET(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDELETE(t *testing.T) {
	setup()
	dataMap["name"] = "john"
	req := httptest.NewRequest(http.MethodDelete, "/del?key=name", nil)
	w := httptest.NewRecorder()
	DELETE(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := dataMap["name"]; ok {
		t.Error("expected key to be deleted")
	}
}
