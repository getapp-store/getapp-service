package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Application{})
}

type Application struct {
	gorm.Model

	Name        string
	Bundle      string
	ApiToken    string
	VkAuthToken string
}
