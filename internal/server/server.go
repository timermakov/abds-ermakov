package server

import (
	"database/sql"
	"net/http"

	"github.com/timermakov/abds-ermakov/internal/config"
	"github.com/timermakov/abds-ermakov/internal/handler/ping"
	"github.com/timermakov/abds-ermakov/internal/httpx"
)

func New(_ config.Config, db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", ping.Handler)
	registerUsers(mux, db)
	return httpx.Recover(mux)
}
