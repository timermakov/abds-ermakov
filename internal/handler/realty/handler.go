package realty

import (
	"context"
	"net/http"

	"github.com/timermakov/abds-ermakov/internal/httpx"
	"github.com/timermakov/abds-ermakov/internal/model"
)

type Finder interface {
	Get(ctx context.Context, id uint64) (model.Realty, bool, error)
}

type realtyResponse struct {
	ID      uint64 `json:"id"`
	Address string `json:"address"`
	Square  uint32 `json:"square"`
	Type    string `json:"type"`
	Number  string `json:"number"`
	Rooms   uint8  `json:"rooms"`
	Floor   uint8  `json:"floor"`
	Porch   uint8  `json:"porch"`
}

type GetHandler struct {
	finder Finder
}

func NewGetHandler(finder Finder) GetHandler {
	return GetHandler{finder: finder}
}

func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	realty, found, err := h.finder.Get(r.Context(), id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		httpx.NotFound(w)
		return
	}
	httpx.JSON(w, http.StatusOK, realtyResponse{
		ID:      realty.ID,
		Address: realty.Address,
		Square:  realty.Square,
		Type:    realty.Type,
		Number:  realty.Number,
		Rooms:   realty.Rooms,
		Floor:   realty.Floor,
		Porch:   realty.Porch,
	})
}
