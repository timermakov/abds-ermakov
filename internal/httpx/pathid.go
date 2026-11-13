package httpx

import (
	"net/http"
	"strconv"
)

func PathID(r *http.Request, name string) (uint64, error) {
	return strconv.ParseUint(r.PathValue(name), 10, 64)
}
