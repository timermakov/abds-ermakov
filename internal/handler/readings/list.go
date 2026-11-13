package readings

import (
	"net/http"
	"time"

	"github.com/timermakov/abds-ermakov/internal/httpx"
)

type readingResponse struct {
	ReadingDate string `json:"reading_date"`
	UserID      uint64 `json:"user_id"`
	TariffID    uint64 `json:"tariff_id"`
	GaugeID     uint64 `json:"gauge_id"`
	DayKW       uint32 `json:"day_kw"`
	NightKW     uint32 `json:"night_kw"`
}

type listResponse struct {
	Readings []readingResponse `json:"readings"`
	Count    int               `json:"count"`
}

type ListHandler struct {
	lister   Lister
	realties RealtyFinder
}

func NewListHandler(lister Lister, realties RealtyFinder) ListHandler {
	return ListHandler{lister: lister, realties: realties}
}

func (h ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	realtyID, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	params := r.URL.Query()
	from, ok := parseDay(params.Get("date_from"))
	if !ok {
		httpx.InvalidParameter(w, "date_from")
		return
	}
	to, ok := parseDay(params.Get("date_to"))
	if !ok {
		httpx.InvalidParameter(w, "date_to")
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
	items, err := h.lister.List(r.Context(), realtyID, from, to)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	readings := make([]readingResponse, 0, len(items))
	for _, item := range items {
		readings = append(readings, readingResponse{
			ReadingDate: item.ReadingDate.Format(dayLayout),
			UserID:      item.UserID,
			TariffID:    item.TariffID,
			GaugeID:     item.GaugeID,
			DayKW:       item.DayKW,
			NightKW:     item.NightKW,
		})
	}
	httpx.JSON(w, http.StatusOK, listResponse{Readings: readings, Count: len(readings)})
}

func parseDay(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	day, err := time.Parse(dayLayout, value)
	if err != nil {
		return time.Time{}, false
	}
	return day, true
}
