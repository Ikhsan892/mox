package models

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// CustomDeletedAt is a custom GORM type to handle Unix epoch timestamps for deleted records
type CustomDeletedAt struct {
	gorm.DeletedAt
}

// MarshalJSON converts the DeletedAt field to a Unix epoch timestamp
func (d CustomDeletedAt) MarshalJSON() ([]byte, error) {
	if d.Valid {
		return []byte(strconv.FormatInt(d.Time.UnixMilli(), 10)), nil
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses a Unix epoch timestamp to DeletedAt
func (d *CustomDeletedAt) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		d.Valid = false
		return nil
	}
	ts, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	d.Time = time.UnixMilli(ts)
	d.Valid = true
	return nil
}

// BaseModel defines the common fields for all database models
type BaseModel struct {
	ID        string `gorm:"type:char(26);primaryKey"`
	CreatedAt time.Time
	Created   int64           `gorm:"autoCreateTime:milli"`
	Updated   int64           `gorm:"autoUpdateTime:milli"`
	DeletedAt CustomDeletedAt `gorm:"index"`
}

// BeforeCreate hook to generate a ULID before inserting a new record
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		return err
	}
	b.ID = id.String()
	return nil
}
