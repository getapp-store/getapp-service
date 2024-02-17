package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Unit{})
}

const (
	UnitFormatInterstitial = "interstitial"
	UnitFormatNative       = "native"
	UnitFormatBanner       = "banner"
	UnitFormatRewarded     = "rewarded"
)

// Unit это представления рекламного юнита из конкретной сетки
type Unit struct {
	gorm.Model

	Name string
	Unit string // unit id from network

	// Format string ?

	NetworkId uint
	Network   Network

	PlacementId uint
	Placement   Placement

	Active bool

	Data string
}
