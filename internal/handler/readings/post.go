package readings

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/timermakov/abds-ermakov/internal/httpx"
	"github.com/timermakov/abds-ermakov/internal/model"
)

const dayLayout = "2006-01-02"

type Creator interface {
	Create(ctx context.Context, reading model.Reading) error
}

type Lister interface {
	List(ctx context.Context, realtyID uint64, from time.Time, to time.Time) ([]model.Reading, error)
}

type RealtyFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type UserFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type GaugeFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type TariffFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type readingRequest struct {
	UserID      *uint64 `json:"user_id"`
	GaugeID     *uint64 `json:"gauge_id"`
	TariffID    *uint64 `json:"tariff_id"`
	ReadingDate *string `json:"reading_date"`
	DayKW       *int64  `json:"day_kw"`
	NightKW     *int64  `json:"night_kw"`
}

type PostHandler struct {
	creator  Creator
	realties RealtyFinder
	users    UserFinder
	gauges   GaugeFinder
	tariffs  TariffFinder
}

func NewPostHandler(creator Creator, realties RealtyFinder, users UserFinder, gauges GaugeFinder, tariffs TariffFinder) PostHandler {
	return PostHandler{
		creator:  creator,
		realties: realties,
		users:    users,
		gauges:   gauges,
		tariffs:  tariffs,
	}
}

func (h PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	realtyID, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	var body readingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.UserID == nil {
		httpx.InvalidField(w, "user_id")
		return
	}
	if body.GaugeID == nil {
		httpx.InvalidField(w, "gauge_id")
		return
	}
	if body.TariffID == nil {
		httpx.InvalidField(w, "tariff_id")
		return
	}
	if body.ReadingDate == nil {
		httpx.InvalidField(w, "reading_date")
		return
	}
	if body.DayKW == nil {
		httpx.InvalidField(w, "day_kw")
		return
	}
	if body.NightKW == nil {
		httpx.InvalidField(w, "night_kw")
		return
	}
	if *body.DayKW < 0 {
		httpx.InvalidField(w, "day_kw")
		return
	}
	if *body.NightKW < 0 {
		httpx.InvalidField(w, "night_kw")
		return
	}
	readingDate, err := time.Parse(dayLayout, *body.ReadingDate)
	if err != nil {
		httpx.InvalidField(w, "reading_date")
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
	userExists, err := h.users.Exists(r.Context(), *body.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !userExists {
		httpx.NotFound(w)
		return
	}
	gaugeExists, err := h.gauges.Exists(r.Context(), *body.GaugeID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !gaugeExists {
		httpx.NotFound(w)
		return
	}
	tariffExists, err := h.tariffs.Exists(r.Context(), *body.TariffID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !tariffExists {
		httpx.NotFound(w)
		return
	}
	reading := model.Reading{
		RealtyID:    realtyID,
		UserID:      *body.UserID,
		TariffID:    *body.TariffID,
		GaugeID:     *body.GaugeID,
		ReadingDate: readingDate,
		DayKW:       uint32(*body.DayKW),
		NightKW:     uint32(*body.NightKW),
		CreatedAt:   time.Now().UTC(),
	}
	if err := h.creator.Create(r.Context(), reading); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.Empty(w, http.StatusCreated)
}
