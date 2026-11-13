package bindings

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

func (r Repo) Active(ctx context.Context, realtyID uint64) (model.Binding, bool, error) {
	const query = "SELECT gauge_id, realty_id, created_at, sign, version FROM fact_gauge_bindings FINAL WHERE realty_id = ? AND sign = 1 LIMIT 1"
	var binding model.Binding
	err := r.db.QueryRowContext(ctx, query, realtyID).Scan(
		&binding.GaugeID,
		&binding.RealtyID,
		&binding.CreatedAt,
		&binding.Sign,
		&binding.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Binding{}, false, nil
	}
	if err != nil {
		return model.Binding{}, false, fmt.Errorf("read active binding of realty %d: %w", realtyID, err)
	}
	return binding, true, nil
}

func (r Repo) Install(ctx context.Context, realtyID uint64, gaugeID uint64) error {
	active, found, err := r.Active(ctx, realtyID)
	if err != nil {
		return err
	}
	rows := InstallRows(active, found, realtyID, gaugeID, time.Now().UTC())
	const query = "INSERT INTO fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) VALUES (?, ?, ?, ?, ?)"
	for _, row := range rows {
		_, err := r.db.ExecContext(
			ctx,
			query,
			row.GaugeID,
			row.RealtyID,
			row.CreatedAt,
			row.Sign,
			row.Version,
		)
		if err != nil {
			return fmt.Errorf("insert binding of realty %d: %w", realtyID, err)
		}
	}
	return nil
}

func InstallRows(active model.Binding, found bool, realtyID uint64, gaugeID uint64, now time.Time) []model.Binding {
	var rows []model.Binding
	version := uint64(1)
	if found {
		rows = append(rows, model.Binding{
			GaugeID:   active.GaugeID,
			RealtyID:  realtyID,
			CreatedAt: now,
			Sign:      -1,
			Version:   active.Version,
		})
		version = active.Version + 1
	}
	rows = append(rows, model.Binding{
		GaugeID:   gaugeID,
		RealtyID:  realtyID,
		CreatedAt: now,
		Sign:      1,
		Version:   version,
	})
	return rows
}
