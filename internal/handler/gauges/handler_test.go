package gauges

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeInstaller struct {
	realtyID uint64
	gaugeID  uint64
	err      error
	calls    int
}

func (f *fakeInstaller) Install(_ context.Context, realtyID uint64, gaugeID uint64) error {
	f.calls++
	f.realtyID = realtyID
	f.gaugeID = gaugeID
	return f.err
}

type fakeExists struct {
	found bool
	err   error
}

func (f fakeExists) Exists(_ context.Context, _ uint64) (bool, error) {
	return f.found, f.err
}

func newRequest(body string, id string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/realty/"+id+"/gauges", strings.NewReader(body))
	req.SetPathValue("id", id)
	return req
}

func TestPostHandlerInstalls(t *testing.T) {
	installer := &fakeInstaller{}
	handler := NewPostHandler(installer, fakeExists{found: true}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": 2}`, "1"))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
	if installer.calls != 1 {
		t.Fatalf("calls = %d", installer.calls)
	}
	if installer.realtyID != 1 || installer.gaugeID != 2 {
		t.Errorf("installed = realty %d gauge %d", installer.realtyID, installer.gaugeID)
	}
}

func TestPostHandlerMissingGauge(t *testing.T) {
	installer := &fakeInstaller{}
	handler := NewPostHandler(installer, fakeExists{found: true}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{}`, "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message != "invalid \"gauge_id\" field" {
		t.Errorf("message = %q", got.Message)
	}
	if installer.calls != 0 {
		t.Errorf("calls = %d, installer must not be called", installer.calls)
	}
}

func TestPostHandlerZeroGauge(t *testing.T) {
	installer := &fakeInstaller{}
	handler := NewPostHandler(installer, fakeExists{found: true}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": 0}`, "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestPostHandlerUnknownRealty(t *testing.T) {
	installer := &fakeInstaller{}
	handler := NewPostHandler(installer, fakeExists{found: false}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": 2}`, "999999999"))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	if installer.calls != 0 {
		t.Errorf("calls = %d", installer.calls)
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

func TestPostHandlerUnknownGauge(t *testing.T) {
	installer := &fakeInstaller{}
	handler := NewPostHandler(installer, fakeExists{found: true}, fakeExists{found: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": 2}`, "1"))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	if installer.calls != 0 {
		t.Errorf("calls = %d", installer.calls)
	}
}

func TestPostHandlerMalformedJSON(t *testing.T) {
	handler := NewPostHandler(&fakeInstaller{}, fakeExists{found: true}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": `, "1"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestPostHandlerInstallerError(t *testing.T) {
	handler := NewPostHandler(&fakeInstaller{err: errors.New("db down")}, fakeExists{found: true}, fakeExists{found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequest(`{"gauge_id": 2}`, "1"))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}
