package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Subscription{})
}

type Subscription struct {
	gorm.Model

	External int
	Name     string
	Title    string
	Amount   int
	Active   bool
	BlogID   uint
	Blog     Blog
}
