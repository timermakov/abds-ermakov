package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorBody struct {
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println(err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, errorBody{Message: message})
}

func Empty(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "Not found")
}

func InvalidField(w http.ResponseWriter, name string) {
	Error(w, http.StatusBadRequest, fmt.Sprintf("invalid %q field", name))
}

func InvalidParameter(w http.ResponseWriter, name string) {
	Error(w, http.StatusBadRequest, fmt.Sprintf("invalid %q parameter", name))
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				Error(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
