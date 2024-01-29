package models

import "gorm.io/gorm"

type Cpa struct {
	gorm.Model

	Url   string
	Image string
}
