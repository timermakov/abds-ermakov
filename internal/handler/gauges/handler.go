package gauges

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/timermakov/abds-ermakov/internal/httpx"
)

type Installer interface {
	Install(ctx context.Context, realtyID uint64, gaugeID uint64) error
}

type RealtyFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type GaugeFinder interface {
	Exists(ctx context.Context, id uint64) (bool, error)
}

type installRequest struct {
	GaugeID *int64 `json:"gauge_id"`
}

type PostHandler struct {
	installer Installer
	realties  RealtyFinder
	gauges    GaugeFinder
}

func NewPostHandler(installer Installer, realties RealtyFinder, gauges GaugeFinder) PostHandler {
	return PostHandler{installer: installer, realties: realties, gauges: gauges}
}

func (h PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	realtyID, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	var body installRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.GaugeID == nil || *body.GaugeID <= 0 {
		httpx.InvalidField(w, "gauge_id")
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
	gaugeExists, err := h.gauges.Exists(r.Context(), uint64(*body.GaugeID))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !gaugeExists {
		httpx.NotFound(w)
		return
	}
	if err := h.installer.Install(r.Context(), realtyID, uint64(*body.GaugeID)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.Empty(w, http.StatusCreated)
}
