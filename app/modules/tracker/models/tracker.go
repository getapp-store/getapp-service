package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Tracker{})
}

type Tracker struct {
	gorm.Model

	Name                 string
	Title                string
	ApplicationID        uint
	Application          applications.Application
	YandexMetricaTracker string
	VkTracker            string
	Active               bool
	YandexToken          string
}
