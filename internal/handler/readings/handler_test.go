package readings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type fakeCreator struct {
	reading model.Reading
	err     error
	calls   int
}

func (f *fakeCreator) Create(_ context.Context, reading model.Reading) error {
	f.calls++
	f.reading = reading
	return f.err
}

type fakeExists struct {
	found bool
	err   error
	calls int
}

func (f *fakeExists) Exists(_ context.Context, _ uint64) (bool, error) {
	f.calls++
	return f.found, f.err
}

type fakeLister struct {
	items []model.Reading
	err   error
	calls int
}

func (f *fakeLister) List(_ context.Context, _ uint64, _ time.Time, _ time.Time) ([]model.Reading, error) {
	f.calls++
	return f.items, f.err
}

func postHandler(creator *fakeCreator, exists bool) PostHandler {
	return NewPostHandler(
		creator,
		&fakeExists{found: exists},
		&fakeExists{found: exists},
		&fakeExists{found: exists},
		&fakeExists{found: exists},
	)
}

func newRequest(method string, target string, body string, id string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.SetPathValue("id", id)
	return req
}

func validBody() string {
	return `{"user_id": 1, "gauge_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "day_kw": 1250, "night_kw": 300}`
}

func TestPostHandlerCreated(t *testing.T) {
	creator := &fakeCreator{}
	handler := postHandler(creator, true)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", validBody(), "1"))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
	if creator.calls != 1 {
		t.Fatalf("calls = %d", creator.calls)
	}
	if creator.reading.RealtyID != 1 {
		t.Errorf("realty_id = %d", creator.reading.RealtyID)
	}
	if creator.reading.UserID != 1 || creator.reading.GaugeID != 1 || creator.reading.TariffID != 1 {
		t.Errorf("references = %d/%d/%d", creator.reading.UserID, creator.reading.GaugeID, creator.reading.TariffID)
	}
	if creator.reading.DayKW != 1250 || creator.reading.NightKW != 300 {
		t.Errorf("counters = %d/%d", creator.reading.DayKW, creator.reading.NightKW)
	}
	if !creator.reading.ReadingDate.Equal(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("reading_date = %v", creator.reading.ReadingDate)
	}
	if creator.reading.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}
}

func TestPostHandlerMissingFields(t *testing.T) {
	bodies := map[string]string{
		"user_id":      `{"gauge_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "day_kw": 1, "night_kw": 1}`,
		"gauge_id":     `{"user_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "day_kw": 1, "night_kw": 1}`,
		"tariff_id":    `{"user_id": 1, "gauge_id": 1, "reading_date": "2026-10-01", "day_kw": 1, "night_kw": 1}`,
		"reading_date": `{"user_id": 1, "gauge_id": 1, "tariff_id": 1, "day_kw": 1, "night_kw": 1}`,
		"day_kw":       `{"user_id": 1, "gauge_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "night_kw": 1}`,
		"night_kw":     `{"user_id": 1, "gauge_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "day_kw": 1}`,
	}
	for field, body := range bodies {
		t.Run(field, func(t *testing.T) {
			handler := postHandler(&fakeCreator{}, true)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", body, "1"))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", recorder.Code)
			}
			var got struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			want := "invalid \"" + field + "\" field"
			if got.Message != want {
				t.Errorf("message = %q, want %q", got.Message, want)
			}
		})
	}
}

func TestPostHandlerNegativeCounters(t *testing.T) {
	for _, field := range []string{"day_kw", "night_kw"} {
		t.Run(field, func(t *testing.T) {
			body := strings.Replace(validBody(), field, "negative_"+field, 1)
			body = strings.Replace(body, "negative_day_kw", "\"day_kw\": -1", 1)
			body = strings.Replace(body, "negative_night_kw", "\"night_kw\": -1", 1)
			handler := postHandler(&fakeCreator{}, true)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", body, "1"))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", recorder.Code)
			}
		})
	}
}

func TestPostHandlerBadDate(t *testing.T) {
	body := strings.Replace(validBody(), "2026-10-01", "not-a-date", 1)
	handler := postHandler(&fakeCreator{}, true)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", body, "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message != "invalid \"reading_date\" field" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestPostHandlerUnknownRealty(t *testing.T) {
	handler := postHandler(&fakeCreator{}, false)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/999/readings", validBody(), "999"))
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

func TestPostHandlerCreatorError(t *testing.T) {
	handler := postHandler(&fakeCreator{err: errors.New("db down")}, true)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", validBody(), "1"))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestPostHandlerMalformedJSON(t *testing.T) {
	handler := postHandler(&fakeCreator{}, true)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodPost, "/realty/1/readings", `{"user_id": `, "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func readingSample() model.Reading {
	return model.Reading{
		RealtyID:    1,
		UserID:      1,
		TariffID:    1,
		GaugeID:     1,
		ReadingDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		DayKW:       1250,
		NightKW:     300,
	}
}

func TestListHandlerReturnsReadings(t *testing.T) {
	lister := &fakeLister{items: []model.Reading{readingSample()}}
	handler := NewListHandler(lister, &fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/1/readings?date_from=2026-10-01&date_to=2026-10-31", "", "1"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got listResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Count != 1 || len(got.Readings) != 1 {
		t.Fatalf("count = %d, readings = %d", got.Count, len(got.Readings))
	}
	if got.Readings[0].ReadingDate != "2026-10-01" {
		t.Errorf("reading_date = %q", got.Readings[0].ReadingDate)
	}
	if got.Readings[0].DayKW != 1250 || got.Readings[0].NightKW != 300 {
		t.Errorf("counters = %d/%d", got.Readings[0].DayKW, got.Readings[0].NightKW)
	}
}

func TestListHandlerEmptyPeriod(t *testing.T) {
	lister := &fakeLister{items: nil}
	handler := NewListHandler(lister, &fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/1/readings?date_from=2026-10-01&date_to=2026-10-31", "", "1"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got listResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Readings == nil || len(got.Readings) != 0 {
		t.Errorf("readings = %v, want empty slice", got.Readings)
	}
	if got.Count != 0 {
		t.Errorf("count = %d", got.Count)
	}
}

func TestListHandlerInvalidParams(t *testing.T) {
	cases := []struct {
		query     string
		parameter string
	}{
		{query: "date_from=not-a-date&date_to=2026-10-31", parameter: "date_from"},
		{query: "date_from=2026-10-01&date_to=not-a-date", parameter: "date_to"},
		{query: "date_to=2026-10-31", parameter: "date_from"},
		{query: "date_from=2026-10-01", parameter: "date_to"},
	}
	for _, c := range cases {
		t.Run(c.parameter, func(t *testing.T) {
			handler := NewListHandler(&fakeLister{}, &fakeExists{found: true})
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/1/readings?"+c.query, "", "1"))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", recorder.Code)
			}
			var got struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			want := "invalid \"" + c.parameter + "\" parameter"
			if got.Message != want {
				t.Errorf("message = %q, want %q", got.Message, want)
			}
		})
	}
}

func TestListHandlerInvertedRange(t *testing.T) {
	lister := &fakeLister{}
	handler := NewListHandler(lister, &fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/1/readings?date_from=2026-10-31&date_to=2026-10-01", "", "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if lister.calls != 0 {
		t.Errorf("lister calls = %d", lister.calls)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message != `"date_from" must not be after "date_to"` {
		t.Errorf("message = %q", got.Message)
	}
}

func TestListHandlerUnknownRealty(t *testing.T) {
	handler := NewListHandler(&fakeLister{}, &fakeExists{found: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/999/readings?date_from=2026-10-01&date_to=2026-10-31", "", "999"))
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

func TestListHandlerListerError(t *testing.T) {
	handler := NewListHandler(&fakeLister{err: errors.New("db down")}, &fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(http.MethodGet, "/realty/1/readings?date_from=2026-10-01&date_to=2026-10-31", "", "1"))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}
