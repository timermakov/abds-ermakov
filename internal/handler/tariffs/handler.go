package tariffs

import (
	"context"
	"net/http"
	"time"

	"github.com/timermakov/abds-ermakov/internal/httpx"
	"github.com/timermakov/abds-ermakov/internal/model"
)

const (
	dayLayout      = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04:05"
)

type ByDateFinder interface {
	GetByDate(ctx context.Context, day time.Time) (model.Tariff, bool, error)
}

type tariffResponse struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	PriceDayKWPH   uint32 `json:"price_day_kwph"`
	PriceNightKWPH uint32 `json:"price_night_kwph"`
	ValidFrom      string `json:"valid_from"`
	ValidTo        string `json:"valid_to"`
}

type GetHandler struct {
	finder ByDateFinder
}

func NewGetHandler(finder ByDateFinder) GetHandler {
	return GetHandler{finder: finder}
}

func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("date")
	day, err := time.Parse(dayLayout, value)
	if value == "" || err != nil {
		httpx.InvalidParameter(w, "date")
		return
	}
	tariff, found, err := h.finder.GetByDate(r.Context(), day)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		httpx.NotFound(w)
		return
	}
	httpx.JSON(w, http.StatusOK, tariffResponse{
		ID:             tariff.ID,
		Name:           tariff.Name,
		PriceDayKWPH:   tariff.PriceDayKWPH,
		PriceNightKWPH: tariff.PriceNightKWPH,
		ValidFrom:      tariff.ValidFrom.Format(dateTimeLayout),
		ValidTo:        tariff.ValidTo.Format(dateTimeLayout),
	})
}
