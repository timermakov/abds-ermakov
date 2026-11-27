package invoices

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

type fakeProvider struct {
	items      []model.Invoice
	count      int64
	err        error
	calls      int
	lastLimit  int64
	lastOffset int64
}

func (f *fakeProvider) List(_ context.Context, _ uint64, limit int64, offset int64) ([]model.Invoice, error) {
	f.calls++
	f.lastLimit = limit
	f.lastOffset = offset
	return f.items, f.err
}

func (f *fakeProvider) Count(_ context.Context, _ uint64) (int64, error) {
	return f.count, f.err
}

type fakeRealtyExists struct {
	found bool
	err   error
}

func (f fakeRealtyExists) Exists(_ context.Context, _ uint64) (bool, error) {
	return f.found, f.err
}

func sampleInvoice(created time.Time) model.Invoice {
	return model.Invoice{
		RealtyID:  1,
		UserID:    1,
		TariffID:  1,
		GaugeID:   1,
		DayKW:     200,
		NightKW:   103,
		Price:     155930,
		Deadline:  time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC),
		CreatedAt: created,
	}
}

func newRequest(target string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.SetPathValue("id", "1")
	return req
}

func TestListHandlerReturnsInvoices(t *testing.T) {
	provider := &fakeProvider{
		items: []model.Invoice{sampleInvoice(time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC))},
		count: 36,
	}
	handler := NewListHandler(provider, fakeRealtyExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest("/realty/1/invoices?limit=10&offset=0"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got listResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Count != 36 {
		t.Errorf("count = %d", got.Count)
	}
	if len(got.Invoices) != 1 {
		t.Fatalf("invoices = %d", len(got.Invoices))
	}
	if got.Invoices[0].Price != 155930 {
		t.Errorf("price = %d", got.Invoices[0].Price)
	}
	if got.Invoices[0].Deadline != "2026-11-10" {
		t.Errorf("deadline = %q", got.Invoices[0].Deadline)
	}
	if got.Invoices[0].CreatedAt != "2026-10-27 10:00:00" {
		t.Errorf("created_at = %q", got.Invoices[0].CreatedAt)
	}
}

func TestListHandlerDefaultsPagination(t *testing.T) {
	provider := &fakeProvider{count: 36}
	handler := NewListHandler(provider, fakeRealtyExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest("/realty/1/invoices"))
	if provider.lastLimit != 10 {
		t.Errorf("limit = %d, want default 10", provider.lastLimit)
	}
	if provider.lastOffset != 0 {
		t.Errorf("offset = %d, want default 0", provider.lastOffset)
	}
}

func TestListHandlerPassesPagination(t *testing.T) {
	provider := &fakeProvider{}
	handler := NewListHandler(provider, fakeRealtyExists{found: true})
	handler.ServeHTTP(httptest.NewRecorder(), newRequest("/realty/1/invoices?limit=5&offset=15"))
	if provider.lastLimit != 5 || provider.lastOffset != 15 {
		t.Errorf("limit/offset = %d/%d", provider.lastLimit, provider.lastOffset)
	}
}

func TestListHandlerInvalidPagination(t *testing.T) {
	cases := []struct {
		query     string
		parameter string
	}{
		{query: "limit=-1&offset=0", parameter: "limit"},
		{query: "limit=abc&offset=0", parameter: "limit"},
		{query: "limit=10&offset=-2", parameter: "offset"},
		{query: "limit=10&offset=x", parameter: "offset"},
	}
	for _, c := range cases {
		t.Run(c.parameter, func(t *testing.T) {
			handler := NewListHandler(&fakeProvider{}, fakeRealtyExists{found: true})
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, newRequest("/realty/1/invoices?"+c.query))
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

func TestListHandlerUnknownRealty(t *testing.T) {
	handler := NewListHandler(&fakeProvider{}, fakeRealtyExists{found: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest("/realty/999/invoices"))
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

func TestListHandlerProviderError(t *testing.T) {
	handler := NewListHandler(&fakeProvider{err: errors.New("db down")}, fakeRealtyExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest("/realty/1/invoices"))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestListHandlerEmptyPage(t *testing.T) {
	provider := &fakeProvider{items: nil, count: 36}
	handler := NewListHandler(provider, fakeRealtyExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest("/realty/1/invoices?limit=0&offset=0"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got listResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Invoices == nil || len(got.Invoices) != 0 {
		t.Errorf("invoices = %v, want empty slice", got.Invoices)
	}
	if got.Count != 36 {
		t.Errorf("count = %d, must ignore limit", got.Count)
	}
}
