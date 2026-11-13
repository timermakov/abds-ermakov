package bindings

import (
	"testing"
	"time"

	"github.com/timermakov/abds-ermakov/internal/model"
)

func activeBinding() model.Binding {
	return model.Binding{
		GaugeID:   1,
		RealtyID:  1,
		CreatedAt: time.Date(2023, 10, 20, 10, 0, 0, 0, time.UTC),
		Sign:      1,
		Version:   1,
	}
}

func TestInstallRowsWithoutActive(t *testing.T) {
	var active model.Binding
	rows := InstallRows(active, false, 1, 2, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if len(rows) != 1 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].GaugeID != 2 || rows[0].RealtyID != 1 || rows[0].Sign != 1 || rows[0].Version != 1 {
		t.Errorf("row = %+v", rows[0])
	}
}

func TestInstallRowsCancelsActiveAndInstalls(t *testing.T) {
	rows := InstallRows(activeBinding(), true, 1, 2, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].GaugeID != 1 || rows[0].Sign != -1 || rows[0].Version != 1 {
		t.Errorf("cancel row = %+v", rows[0])
	}
	if rows[0].GaugeID != activeBinding().GaugeID {
		t.Error("cancel row must reference the canceled gauge")
	}
	if rows[1].GaugeID != 2 || rows[1].Sign != 1 || rows[1].Version != 2 {
		t.Errorf("install row = %+v", rows[1])
	}
}

func TestInstallRowsIncrementsVersion(t *testing.T) {
	active := activeBinding()
	active.Version = 5
	rows := InstallRows(active, true, 1, 3, time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC))
	if rows[0].Version != 5 {
		t.Errorf("cancel version = %d, must match active", rows[0].Version)
	}
	if rows[1].Version != 6 {
		t.Errorf("install version = %d, must be active + 1", rows[1].Version)
	}
}
