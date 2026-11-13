package ping

import (
	"net/http/httptest"
	"testing"
)

func TestHandlerWritesEmptyOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	Handler(recorder, nil)
	if recorder.Code != 200 {
		t.Errorf("status = %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
}
