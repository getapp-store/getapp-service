package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Pincode{})
}

type Pincode struct {
	gorm.Model

	Code   string
	Ttl    int
	UserID uint
	User   User
}
