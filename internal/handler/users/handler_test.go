package users

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type fakeFinder struct {
	user  model.User
	found bool
	err   error
}

func (f fakeFinder) Get(_ context.Context, _ uint64) (model.User, bool, error) {
	return f.user, f.found, f.err
}

type fakePatcher struct {
	patch model.UserPatch
	found bool
	err   error
	calls int
}

func (f *fakePatcher) Patch(_ context.Context, _ uint64, patch model.UserPatch) (bool, error) {
	f.calls++
	f.patch = patch
	return f.found, f.err
}

func sampleUser() model.User {
	return model.User{
		ID:        1,
		FIO:       "Иванов Иван Иванович",
		Phone:     "+79000000001",
		PhonePrev: "",
	}
}

func request(method string, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.SetPathValue("id", "1")
	return req
}

func TestGetHandlerFound(t *testing.T) {
	handler := NewGetHandler(fakeFinder{user: sampleUser(), found: true})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodGet, "/users/1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got userResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("id = %d", got.ID)
	}
	if got.FIO != "Иванов Иван Иванович" {
		t.Errorf("fio = %q", got.FIO)
	}
	if got.Phone != "+79000000001" {
		t.Errorf("phone = %q", got.Phone)
	}
	if got.PhonePrev != "" {
		t.Errorf("phone_prev = %q", got.PhonePrev)
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	handler := NewGetHandler(fakeFinder{found: false})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodGet, "/users/1", nil))
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
	handler.ServeHTTP(recorder, request(http.MethodGet, "/users/1", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestGetHandlerBadID(t *testing.T) {
	handler := NewGetHandler(fakeFinder{})
	recorder := httptest.NewRecorder()
	req := request(http.MethodGet, "/users/abc", nil)
	req.SetPathValue("id", "abc")
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestPatchHandlerAppliesPatch(t *testing.T) {
	patcher := &fakePatcher{found: true}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"fio": "Петрова Анна Сергеевна", "phone": "+79111111111"}`)
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", body))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
	if patcher.calls != 1 {
		t.Fatalf("calls = %d", patcher.calls)
	}
	if patcher.patch.FIO == nil || *patcher.patch.FIO != "Петрова Анна Сергеевна" {
		t.Errorf("patch fio = %v", patcher.patch.FIO)
	}
	if patcher.patch.Phone == nil || *patcher.patch.Phone != "+79111111111" {
		t.Errorf("patch phone = %v", patcher.patch.Phone)
	}
}

func TestPatchHandlerNotFound(t *testing.T) {
	patcher := &fakePatcher{found: false}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", strings.NewReader(`{"phone": "+79111111111"}`)))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	if patcher.calls != 1 {
		t.Errorf("calls = %d, patcher must decide the not found case", patcher.calls)
	}
}

func TestPatchHandlerEmptyBody(t *testing.T) {
	patcher := &fakePatcher{found: true}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", strings.NewReader(`{}`)))
	if recorder.Code != http.StatusBadRequest {
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

func TestPatchHandlerEmptyFio(t *testing.T) {
	patcher := &fakePatcher{found: true}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", strings.NewReader(`{"fio": ""}`)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Message != "invalid \"fio\" field" {
		t.Errorf("message = %q", got.Message)
	}
}

func TestPatchHandlerMalformedJSON(t *testing.T) {
	patcher := &fakePatcher{found: true}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", strings.NewReader(`{"fio": `)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestPatchHandlerRepoError(t *testing.T) {
	patcher := &fakePatcher{err: errors.New("db down")}
	handler := NewPatchHandler(patcher)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request(http.MethodPatch, "/users/1", strings.NewReader(`{"phone": "+79111111111"}`)))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}
