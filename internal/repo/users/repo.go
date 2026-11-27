package users

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

func (r Repo) Get(ctx context.Context, id uint64) (model.User, bool, error) {
	const query = "SELECT id, fio, phone, phone_prev, created_at, updated_at FROM dim_users FINAL WHERE id = ?"
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FIO,
		&user.Phone,
		&user.PhonePrev,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, fmt.Errorf("get user %d: %w", id, err)
	}
	return user, true, nil
}

func (r Repo) Exists(ctx context.Context, id uint64) (bool, error) {
	_, found, err := r.Get(ctx, id)
	return found, err
}

func (r Repo) Patch(ctx context.Context, id uint64, patch model.UserPatch) (bool, error) {
	user, found, err := r.Get(ctx, id)
	if err != nil || !found {
		return found, err
	}
	updated := NextState(user, patch, time.Now().UTC())
	const query = "INSERT INTO dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = r.db.ExecContext(
		ctx,
		query,
		updated.ID,
		updated.FIO,
		updated.Phone,
		updated.PhonePrev,
		updated.CreatedAt,
		updated.UpdatedAt,
	)
	if err != nil {
		return true, fmt.Errorf("patch user %d: %w", id, err)
	}
	return true, nil
}

func NextState(user model.User, patch model.UserPatch, now time.Time) model.User {
	updated := user
	if patch.FIO != nil {
		updated.FIO = *patch.FIO
	}
	if patch.Phone != nil && *patch.Phone != user.Phone {
		updated.Phone = *patch.Phone
		updated.PhonePrev = user.Phone
	}
	updated.UpdatedAt = NextUpdatedAt(user.UpdatedAt, now)
	return updated
}

func NextUpdatedAt(prev time.Time, now time.Time) time.Time {
	if !now.After(prev) {
		return prev.Add(time.Second)
	}
	return now
}
