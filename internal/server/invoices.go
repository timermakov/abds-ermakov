package server

import (
	"database/sql"
	"net/http"

	invoiceshandler "github.com/timermakov/abds-ermakov/internal/handler/invoices"
	invoicesrepo "github.com/timermakov/abds-ermakov/internal/repo/invoices"
	realtyrepo "github.com/timermakov/abds-ermakov/internal/repo/realty"
)

func registerInvoices(mux *http.ServeMux, db *sql.DB) {
	repo := invoicesrepo.NewRepo(db)
	mux.Handle("GET /realty/{id}/invoices", invoiceshandler.NewListHandler(repo, realtyrepo.NewRepo(db)))
}
