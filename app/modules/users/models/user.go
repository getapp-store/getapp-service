package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, User{})
}

type User struct {
	gorm.Model

	ExternalId    int
	VkAccessToken string
	Email         string
	ApplicationID uint
	Application   applications.Application
	ApiToken      string
}
