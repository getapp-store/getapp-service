package models

import (
	"database/sql"
	"time"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Cpm{})
}

type Cpm struct {
	PlacementId uint `gorm:"primaryKey"`
	Placement   Placement

	NetworkId uint `gorm:"primaryKey"`
	Network   Network

	UnitId uint `gorm:"primaryKey"`
	Unit   Unit

	Date time.Time `gorm:"primaryKey"`

	Amount float64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime `gorm:"index"`
}
