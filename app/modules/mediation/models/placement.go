package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Placement{})
}

// Placement объединяет юниты из внешних сеток в одном плейсменте
type Placement struct {
	gorm.Model

	Name   string
	Format string

	ApplicationId uint
	Application   applications.Application

	Units []Unit

	Active bool
}
