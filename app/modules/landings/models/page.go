package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Page{})
}

type Page struct {
	gorm.Model

	Name      string
	Path      string
	Body      string
	LandingID uint
	Landing   Landing
	Active    bool
}
