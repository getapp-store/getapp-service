package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Setting{})
}

type Setting struct {
	gorm.Model

	Key   string
	Value string
}
