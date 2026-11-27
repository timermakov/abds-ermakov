package server

import (
	"database/sql"
	"net/http"

	realtyhandler "github.com/timermakov/abds-ermakov/internal/handler/realty"
	realtyrepo "github.com/timermakov/abds-ermakov/internal/repo/realty"
)

func registerRealty(mux *http.ServeMux, db *sql.DB) {
	repo := realtyrepo.NewRepo(db)
	mux.Handle("GET /realty/{id}", realtyhandler.NewGetHandler(repo))
}
