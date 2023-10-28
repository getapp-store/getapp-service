package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Phrase{})
}

type Phrase struct {
	gorm.Model
	Key        string
	Value      string
	LanguageID uint
	Language   Language
	Active     bool
}
