package models

import (
	"time"

	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Impression{})
}

type Impression struct {
	gorm.Model

	PlacementId uint
	Placement   Placement

	NetworkId uint
	Network   Network

	UnitId uint
	Unit   Unit

	Date time.Time

	Revenue float64
	Raw     string
}
