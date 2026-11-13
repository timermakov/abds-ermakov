package realty

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return Repo{db: db}
}

func (r Repo) Get(ctx context.Context, id uint64) (model.Realty, bool, error) {
	const query = "SELECT id, address, square, type, number, rooms, floor, porch, created_at FROM dim_realty WHERE id = ?"
	var realty model.Realty
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&realty.ID,
		&realty.Address,
		&realty.Square,
		&realty.Type,
		&realty.Number,
		&realty.Rooms,
		&realty.Floor,
		&realty.Porch,
		&realty.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Realty{}, false, nil
	}
	if err != nil {
		return model.Realty{}, false, fmt.Errorf("get realty %d: %w", id, err)
	}
	return realty, true, nil
}

func (r Repo) Exists(ctx context.Context, id uint64) (bool, error) {
	_, found, err := r.Get(ctx, id)
	return found, err
}
