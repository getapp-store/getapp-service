package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Blog{})
}

type Blog struct {
	gorm.Model

	Title         string
	Name          string
	Url           string
	ApplicationID uint
	Application   models.Application
	Active        bool
	Token         string
}
