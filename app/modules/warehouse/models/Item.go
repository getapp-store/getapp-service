package models

import (
	"gorm.io/gorm"
	"ru/kovardin/getapp/pkg/database"

	applications "ru/kovardin/getapp/app/modules/applications/models"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Item{})
}

type Item struct {
	gorm.Model

	ApplicationID uint
	Application   applications.Application

	Key    string `gorm:"uniqueIndex"`
	Value  string
	Active bool
}
