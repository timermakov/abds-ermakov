package readings

import (
	"context"
	"database/sql"
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

func (r Repo) Create(ctx context.Context, reading model.Reading) error {
	const query = "INSERT INTO fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(
		ctx,
		query,
		reading.RealtyID,
		reading.UserID,
		reading.TariffID,
		reading.GaugeID,
		reading.ReadingDate,
		reading.DayKW,
		reading.NightKW,
		reading.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create reading: %w", err)
	}
	return nil
}

func (r Repo) List(ctx context.Context, realtyID uint64, from time.Time, to time.Time) ([]model.Reading, error) {
	const query = "SELECT realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at FROM fact_meter_readings FINAL WHERE realty_id = ? AND reading_date >= ? AND reading_date <= ? ORDER BY reading_date"
	rows, err := r.db.QueryContext(ctx, query, realtyID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	var result []model.Reading
	for rows.Next() {
		var reading model.Reading
		if err := rows.Scan(
			&reading.RealtyID,
			&reading.UserID,
			&reading.TariffID,
			&reading.GaugeID,
			&reading.ReadingDate,
			&reading.DayKW,
			&reading.NightKW,
			&reading.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		result = append(result, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate readings: %w", err)
	}
	return result, nil
}
