package ping

import (
	"net/http"

	"github.com/timermakov/abds-ermakov/internal/httpx"
)

func Handler(w http.ResponseWriter, _ *http.Request) {
	httpx.Empty(w, http.StatusOK)
}
