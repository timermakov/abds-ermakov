package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorWritesMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	Error(recorder, http.StatusBadRequest, "something broke")
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("content type = %q", contentType)
	}
	var got errorBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "something broke" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestNotFoundWritesDefaultMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	NotFound(recorder)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d", recorder.Code)
	}
	var got errorBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "Not found" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestInvalidFieldFormatsMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	InvalidField(recorder, "day_kw")
	var got errorBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "invalid \"day_kw\" field" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestInvalidParameterFormatsMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	InvalidParameter(recorder, "date_from")
	var got errorBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "invalid \"date_from\" parameter" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestEmptyWritesNoBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	Empty(recorder, http.StatusCreated)
	if recorder.Code != http.StatusCreated {
		t.Errorf("status = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
}

func TestRecoversPanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})
	server := httptest.NewServer(Recover(panicHandler))
	defer server.Close()
	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d", response.StatusCode)
	}
	var got errorBody
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "internal error" {
		t.Errorf("message = %q", got.Message)
	}
}
