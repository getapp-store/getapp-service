package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Auth{})
}

type Auth struct {
	gorm.Model

	Title         string
	Name          string
	ApplicationID uint
	Application   applications.Application
	Active        bool
	Config        string
}
