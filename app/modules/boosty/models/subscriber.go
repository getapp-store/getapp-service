package models

import (
	"gorm.io/gorm"

	"ru/kovardin/getapp/pkg/database"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Subscriber{})
}

type Subscriber struct {
	gorm.Model

	External       int
	Name           string
	Email          string
	Active         bool
	SubscriptionID uint
	Subscription   Subscription
	BlogID         uint
	Blog           Blog
	Amount         int
}
