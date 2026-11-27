package tariffs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return Repo{db: db}
}

func (r Repo) Exists(ctx context.Context, id uint64) (bool, error) {
	var count uint64
	err := r.db.QueryRowContext(ctx, "SELECT count() FROM dim_tariffs WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check tariff %d: %w", id, err)
	}
	return count > 0, nil
}

func (r Repo) GetByDate(ctx context.Context, day time.Time) (model.Tariff, bool, error) {
	const query = "SELECT id, name, price_day_kwph, price_night_kwph, valid_from, valid_to FROM dim_tariffs FINAL WHERE valid_from <= ? AND valid_to > ? ORDER BY valid_from DESC LIMIT 1"
	var tariff model.Tariff
	err := r.db.QueryRowContext(ctx, query, day, day).Scan(
		&tariff.ID,
		&tariff.Name,
		&tariff.PriceDayKWPH,
		&tariff.PriceNightKWPH,
		&tariff.ValidFrom,
		&tariff.ValidTo,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Tariff{}, false, nil
	}
	if err != nil {
		return model.Tariff{}, false, fmt.Errorf("get tariff on %s: %w", day.Format("2006-01-02"), err)
	}
	return tariff, true, nil
}
