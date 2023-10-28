package models

import (
	"gorm.io/gorm"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	users "ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Payment{})
}

const (
	PaymentStatusCreated = "created"
	PaymentStatusConfirm = "confirm"
	PaymentStatusSuccess = "success"
)

type Payment struct {
	gorm.Model

	Amount        int
	Status        string
	ProductID     uint
	Product       Product
	UserID        uint
	User          users.User
	ApplicationID uint
	Application   applications.Application
}
