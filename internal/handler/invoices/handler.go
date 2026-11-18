package invoices

import (
	"context"
	"net/http"
	"strconv"

	"github.com/timermakov/abds-ermakov/internal/httpx"
	"github.com/timermakov/abds-ermakov/internal/model"
)

const (
	dayLayout      = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04:05"
	defaultLimit   = int64(10)
)

type Provider interface {
	List(ctx context.Context, realtyID uint64, limit int64, offset int64) ([]model.Invoice, error)
	Count(ctx context.Context, realtyID uint64) (int64, error)
}

type RealtyFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type invoiceResponse struct {
	TariffID  uint64 `json:"tariff_id"`
	GaugeID   uint64 `json:"gauge_id"`
	DayKW     uint32 `json:"day_kw"`
	NightKW   uint32 `json:"night_kw"`
	Price     uint64 `json:"price"`
	Deadline  string `json:"deadline"`
	CreatedAt string `json:"created_at"`
}

type listResponse struct {
	Invoices []invoiceResponse `json:"invoices"`
	Count    int64             `json:"count"`
}

type ListHandler struct {
	provider Provider
	realties RealtyFinder
}

func NewListHandler(provider Provider, realties RealtyFinder) ListHandler {
	return ListHandler{provider: provider, realties: realties}
}

func (h ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	realtyID, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	params := r.URL.Query()
	limit, ok := pagination(params.Get("limit"), defaultLimit)
	if !ok {
		httpx.InvalidParameter(w, "limit")
		return
	}
	offset, ok := pagination(params.Get("offset"), 0)
	if !ok {
		httpx.InvalidParameter(w, "offset")
		return
	}
	exists, err := h.realties.Exists(r.Context(), realtyID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !exists {
		httpx.NotFound(w)
		return
	}
	items, err := h.provider.List(r.Context(), realtyID, limit, offset)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	count, err := h.provider.Count(r.Context(), realtyID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	invoices := make([]invoiceResponse, 0, len(items))
	for _, item := range items {
		invoices = append(invoices, invoiceResponse{
			TariffID:  item.TariffID,
			GaugeID:   item.GaugeID,
			DayKW:     item.DayKW,
			NightKW:   item.NightKW,
			Price:     item.Price,
			Deadline:  item.Deadline.Format(dayLayout),
			CreatedAt: item.CreatedAt.Format(dateTimeLayout),
		})
	}
	httpx.JSON(w, http.StatusOK, listResponse{Invoices: invoices, Count: count})
}

func pagination(value string, fallback int64) (int64, bool) {
	if value == "" {
		return fallback, true
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, false
	}
	return parsed, true
}
