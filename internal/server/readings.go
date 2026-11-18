package server

import (
	"database/sql"
	"net/http"

	readingshandler "github.com/timermakov/abds-ermakov/internal/handler/readings"
	gaugesrepo "github.com/timermakov/abds-ermakov/internal/repo/gauges"
	readingsrepo "github.com/timermakov/abds-ermakov/internal/repo/readings"
	realtyrepo "github.com/timermakov/abds-ermakov/internal/repo/realty"
	tariffsrepo "github.com/timermakov/abds-ermakov/internal/repo/tariffs"
	usersrepo "github.com/timermakov/abds-ermakov/internal/repo/users"
)

func registerReadings(mux *http.ServeMux, db *sql.DB) {
	readings := readingsrepo.NewRepo(db)
	realties := realtyrepo.NewRepo(db)
	users := usersrepo.NewRepo(db)
	gauges := gaugesrepo.NewRepo(db)
	tariffs := tariffsrepo.NewRepo(db)
	mux.Handle("POST /realty/{id}/readings", readingshandler.NewPostHandler(readings, realties, users, gauges, tariffs))
	mux.Handle("GET /realty/{id}/readings", readingshandler.NewListHandler(readings, realties))
}
