package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Language{})
}

type Language struct {
	gorm.Model

	Name          string
	Locale        string
	ApplicationID uint
	Application   applications.Application
	Active        bool
}
