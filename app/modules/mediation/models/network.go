package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Network{})
}

type Network struct {
	gorm.Model

	Name string
	Key  string

	ApplicationId uint
	Application   applications.Application

	Active bool
}
