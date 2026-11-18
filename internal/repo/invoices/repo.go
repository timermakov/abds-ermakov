package invoices

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/timermakov/abds-ermakov/internal/model"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return Repo{db: db}
}

func (r Repo) List(ctx context.Context, realtyID uint64, limit int64, offset int64) ([]model.Invoice, error) {
	const query = "SELECT realty_id, user_id, tariff_id, gauge_id, day_kw, night_kw, price, deadline, created_at FROM fact_invoice WHERE realty_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, realtyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	var result []model.Invoice
	for rows.Next() {
		var invoice model.Invoice
		if err := rows.Scan(
			&invoice.RealtyID,
			&invoice.UserID,
			&invoice.TariffID,
			&invoice.GaugeID,
			&invoice.DayKW,
			&invoice.NightKW,
			&invoice.Price,
			&invoice.Deadline,
			&invoice.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		result = append(result, invoice)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invoices: %w", err)
	}
	return result, nil
}

func (r Repo) Count(ctx context.Context, realtyID uint64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT count() FROM fact_invoice WHERE realty_id = ?", realtyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count invoices: %w", err)
	}
	return count, nil
}
