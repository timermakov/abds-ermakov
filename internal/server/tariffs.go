package server

import (
	"database/sql"
	"net/http"

	tariffshandler "github.com/timermakov/abds-ermakov/internal/handler/tariffs"
	tariffsrepo "github.com/timermakov/abds-ermakov/internal/repo/tariffs"
)

func registerTariffs(mux *http.ServeMux, db *sql.DB) {
	repo := tariffsrepo.NewRepo(db)
	mux.Handle("GET /tariffs", tariffshandler.NewGetHandler(repo))
}
