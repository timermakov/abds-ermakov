package tariffs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type fakeFinder struct {
	tariff model.Tariff
	found  bool
	err    error
	calls  int
}

func (f *fakeFinder) GetByDate(_ context.Context, _ time.Time) (model.Tariff, bool, error) {
	f.calls++
	return f.tariff, f.found, f.err
}

func sampleTariff() model.Tariff {
	return model.Tariff{
		ID:             1,
		Name:           "День-Ночь",
		PriceDayKWPH:   620,
		PriceNightKWPH: 310,
		ValidFrom:      time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		ValidTo:        time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestGetHandlerReturnsTariff(t *testing.T) {
	finder := &fakeFinder{tariff: sampleTariff(), found: true}
	handler := NewGetHandler(finder)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tariffs?date=2026-10-27", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got tariffResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("id = %d", got.ID)
	}
	if got.Name != "День-Ночь" {
		t.Errorf("name = %q", got.Name)
	}
	if got.PriceDayKWPH != 620 || got.PriceNightKWPH != 310 {
		t.Errorf("prices = %d/%d", got.PriceDayKWPH, got.PriceNightKWPH)
	}
	if got.ValidFrom != "2023-11-01 00:00:00" {
		t.Errorf("valid_from = %q", got.ValidFrom)
	}
	if got.ValidTo != "2099-01-01 00:00:00" {
		t.Errorf("valid_to = %q", got.ValidTo)
	}
}

func TestGetHandlerMissingDate(t *testing.T) {
	finder := &fakeFinder{found: true}
	handler := NewGetHandler(finder)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tariffs", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if finder.calls != 0 {
		t.Errorf("calls = %d", finder.calls)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message != "invalid \"date\" parameter" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestGetHandlerMalformedDate(t *testing.T) {
	finder := &fakeFinder{found: true}
	handler := NewGetHandler(finder)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tariffs?date=not-a-date", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if finder.calls != 0 {
		t.Errorf("calls = %d", finder.calls)
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

func TestGetHandlerNotFound(t *testing.T) {
	finder := &fakeFinder{found: false}
	handler := NewGetHandler(finder)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tariffs?date=1900-01-01", nil))
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
	handler := NewGetHandler(&fakeFinder{err: errors.New("db down")})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tariffs?date=2026-10-27", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}
