package models

import (
	"github.com/qor5/x/login"
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Admin{})
}

type Admin struct {
	gorm.Model

	login.UserPass
}
