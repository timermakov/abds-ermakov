package gauges

import (
	"context"
	"database/sql"
	"fmt"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return Repo{db: db}
}

func (r Repo) Exists(ctx context.Context, id uint64) (bool, error) {
	var count uint64
	err := r.db.QueryRowContext(ctx, "SELECT count() FROM dim_gauges WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check gauge %d: %w", id, err)
	}
	return count > 0, nil
}
