package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/timermakov/abds-ermakov/internal/httpx"
	"github.com/timermakov/abds-ermakov/internal/model"
)

type Finder interface {
	Get(ctx context.Context, id uint64) (model.User, bool, error)
}

type Patcher interface {
	Patch(ctx context.Context, id uint64, patch model.UserPatch) (bool, error)
}

type userResponse struct {
	ID        uint64 `json:"id"`
	FIO       string `json:"fio"`
	Phone     string `json:"phone"`
	PhonePrev string `json:"phone_prev"`
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
	user, found, err := h.finder.Get(r.Context(), id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		httpx.NotFound(w)
		return
	}
	httpx.JSON(w, http.StatusOK, userResponse{
		ID:        user.ID,
		FIO:       user.FIO,
		Phone:     user.Phone,
		PhonePrev: user.PhonePrev,
	})
}

type patchRequest struct {
	FIO   *string `json:"fio"`
	Phone *string `json:"phone"`
}

type PatchHandler struct {
	patcher Patcher
}

func NewPatchHandler(patcher Patcher) PatchHandler {
	return PatchHandler{patcher: patcher}
}

func (h PatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.InvalidParameter(w, "id")
		return
	}
	var body patchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FIO == nil && body.Phone == nil {
		httpx.Error(w, http.StatusBadRequest, "at least one field is required")
		return
	}
	if body.FIO != nil && *body.FIO == "" {
		httpx.InvalidField(w, "fio")
		return
	}
	if body.Phone != nil && *body.Phone == "" {
		httpx.InvalidField(w, "phone")
		return
	}
	patch := model.UserPatch{FIO: body.FIO, Phone: body.Phone}
	found, err := h.patcher.Patch(r.Context(), id, patch)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		httpx.NotFound(w)
		return
	}
	httpx.Empty(w, http.StatusNoContent)
}
