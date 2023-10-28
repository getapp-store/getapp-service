package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Landing{})
}

type Landing struct {
	gorm.Model

	Name          string
	Path          string
	ApplicationID uint
	Application   applications.Application
	Active        bool
}
