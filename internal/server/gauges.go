package server

import (
	"database/sql"
	"net/http"

	gaugeshandler "github.com/timermakov/abds-ermakov/internal/handler/gauges"
	bindingsrepo "github.com/timermakov/abds-ermakov/internal/repo/bindings"
	gaugesrepo "github.com/timermakov/abds-ermakov/internal/repo/gauges"
	realtyrepo "github.com/timermakov/abds-ermakov/internal/repo/realty"
)

func registerGauges(mux *http.ServeMux, db *sql.DB) {
	repo := bindingsrepo.NewRepo(db)
	mux.Handle("POST /realty/{id}/gauges", gaugeshandler.NewPostHandler(repo, realtyrepo.NewRepo(db), gaugesrepo.NewRepo(db)))
}
