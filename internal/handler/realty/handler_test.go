package realty

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type fakeFinder struct {
	realty model.Realty
	found  bool
	err    error
}

func (f fakeFinder) Get(_ context.Context, _ uint64) (model.Realty, bool, error) {
	return f.realty, f.found, f.err
}

func sampleRealty() model.Realty {
	return model.Realty{
		ID:      1,
		Address: "г. Санкт-Петербург, улица Загородная, дом 1, кв. 1",
		Square:  30,
		Type:    "flat",
		Number:  "1",
		Rooms:   1,
		Floor:   1,
		Porch:   1,
	}
}

func request() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/realty/1", nil)
	req.SetPathValue("id", "1")
	return req
}

func TestGetHandlerFound(t *testing.T) {
	handler := NewGetHandler(fakeFinder{realty: sampleRealty(), found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request())
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got realtyResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("id = %d", got.ID)
	}
	if got.Address == "" {
		t.Error("address is empty")
	}
	if got.Square != 30 {
		t.Errorf("square = %d", got.Square)
	}
	if got.Type != "flat" {
		t.Errorf("type = %q", got.Type)
	}
	if got.Number != "1" {
		t.Errorf("number = %q", got.Number)
	}
	if got.Rooms != 1 || got.Floor != 1 || got.Porch != 1 {
		t.Errorf("rooms/floor/porch = %d/%d/%d", got.Rooms, got.Floor, got.Porch)
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	handler := NewGetHandler(fakeFinder{found: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request())
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message == "" {
		t.Error("message is empty")
	}
}

func TestGetHandlerFinderError(t *testing.T) {
	handler := NewGetHandler(fakeFinder{err: errors.New("db down")})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request())
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestGetHandlerBadID(t *testing.T) {
	handler := NewGetHandler(fakeFinder{})
	recorder := httptest.NewRecorder()
	req := request()
	req.SetPathValue("id", "abc")
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}
