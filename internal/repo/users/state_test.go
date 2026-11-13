package users

import (
	"testing"
	"time"

	"github.com/timermakov/abds-ermakov/internal/model"
)

func baseUser() model.User {
	return model.User{
		ID:         1,
		FIO:        "Иванов Иван Иванович",
		Phone:      "+79000000001",
		PhonePrev:  "",
		CreatedAt:  time.Date(2023, 11, 1, 9, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2023, 11, 1, 9, 0, 0, 0, time.UTC),
	}
}

func str(value string) *string {
	return &value
}

func TestNextStateAppliesFio(t *testing.T) {
	patch := model.UserPatch{FIO: str("Петрова Анна Сергеевна")}
	updated := NextState(baseUser(), patch, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if updated.FIO != "Петрова Анна Сергеевна" {
		t.Errorf("fio = %q", updated.FIO)
	}
	if updated.Phone != "+79000000001" {
		t.Errorf("phone = %q", updated.Phone)
	}
	if updated.PhonePrev != "" {
		t.Errorf("phone_prev = %q", updated.PhonePrev)
	}
}

func TestNextStateAppliesPhone(t *testing.T) {
	patch := model.UserPatch{Phone: str("+79111111111")}
	updated := NextState(baseUser(), patch, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if updated.Phone != "+79111111111" {
		t.Errorf("phone = %q", updated.Phone)
	}
	if updated.PhonePrev != "+79000000001" {
		t.Errorf("phone_prev = %q", updated.PhonePrev)
	}
	if updated.FIO != "Иванов Иван Иванович" {
		t.Errorf("fio = %q", updated.FIO)
	}
}

func TestNextStateKeepsPrevOnSamePhone(t *testing.T) {
	user := baseUser()
	user.PhonePrev = "+79000000000"
	patch := model.UserPatch{Phone: str("+79000000001")}
	updated := NextState(user, patch, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if updated.PhonePrev != "+79000000000" {
		t.Errorf("phone_prev = %q", updated.PhonePrev)
	}
}

func TestNextStateAppliesBothFields(t *testing.T) {
	patch := model.UserPatch{
		FIO:   str("Новое ФИО"),
		Phone: str("+79222222222"),
	}
	updated := NextState(baseUser(), patch, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if updated.FIO != "Новое ФИО" {
		t.Errorf("fio = %q", updated.FIO)
	}
	if updated.Phone != "+79222222222" {
		t.Errorf("phone = %q", updated.Phone)
	}
	if updated.PhonePrev != "+79000000001" {
		t.Errorf("phone_prev = %q", updated.PhonePrev)
	}
}

func TestNextStateKeepsCreatedAt(t *testing.T) {
	updated := NextState(baseUser(), model.UserPatch{FIO: str("X")}, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if !updated.CreatedAt.Equal(baseUser().CreatedAt) {
		t.Errorf("created_at = %v", updated.CreatedAt)
	}
}

func TestNextUpdatedAtUsesNow(t *testing.T) {
	prev := time.Date(2023, 11, 1, 9, 0, 0, 0, time.UTC)
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	got := NextUpdatedAt(prev, now)
	if !got.Equal(now) {
		t.Errorf("updated_at = %v, want %v", got, now)
	}
}

func TestNextUpdatedAtBumpsOnSameSecond(t *testing.T) {
	same := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	got := NextUpdatedAt(same, same)
	want := same.Add(time.Second)
	if !got.Equal(want) {
		t.Errorf("updated_at = %v, want %v", got, want)
	}
}

func TestNextUpdatedAtBumpsOnOlderNow(t *testing.T) {
	prev := time.Date(2026, 11, 13, 10, 0, 5, 0, time.UTC)
	now := prev.Add(-2 * time.Second)
	got := NextUpdatedAt(prev, now)
	if !got.After(prev) {
		t.Errorf("updated_at = %v, must be after %v", got, prev)
	}
}
