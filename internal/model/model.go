package model

import "time"

type User struct {
	ID         uint64
	FIO        string
	Phone      string
	PhonePrev  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UserPatch struct {
	FIO   *string
	Phone *string
}

type Realty struct {
	ID        uint64
	Address   string
	Square    uint32
	Type      string
	Number    string
	Rooms     uint8
	Floor     uint8
	Porch     uint8
	CreatedAt time.Time
}

type Reading struct {
	RealtyID    uint64
	UserID      uint64
	TariffID    uint64
	GaugeID     uint64
	ReadingDate time.Time
	DayKW       uint32
	NightKW     uint32
	CreatedAt   time.Time
}

type Invoice struct {
	RealtyID  uint64
	UserID    uint64
	TariffID  uint64
	GaugeID   uint64
	DayKW     uint32
	NightKW   uint32
	Price     uint64
	Deadline  time.Time
	CreatedAt time.Time
}

type Tariff struct {
	ID             uint64
	Name           string
	PriceDayKWPH   uint32
	PriceNightKWPH uint32
	ValidFrom      time.Time
	ValidTo        time.Time
}

type Binding struct {
	GaugeID   uint64
	RealtyID  uint64
	CreatedAt time.Time
	Sign      int8
	Version   uint64
}
