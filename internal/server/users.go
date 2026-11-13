package server

import (
	"database/sql"
	"net/http"

	usershandler "github.com/timermakov/abds-ermakov/internal/handler/users"
	usersrepo "github.com/timermakov/abds-ermakov/internal/repo/users"
)

func registerUsers(mux *http.ServeMux, db *sql.DB) {
	repo := usersrepo.NewRepo(db)
	mux.Handle("GET /users/{id}", usershandler.NewGetHandler(repo))
	mux.Handle("PATCH /users/{id}", usershandler.NewPatchHandler(repo))
}
