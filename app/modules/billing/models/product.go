package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Product{})
}

type Product struct {
	gorm.Model

	Name          string `gorm:"index:products_name,unique"`
	Title         string
	Amount        int
	ApplicationID uint
	Application   applications.Application
	Active        bool
}
