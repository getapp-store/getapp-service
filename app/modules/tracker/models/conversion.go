package models

import (
	"gorm.io/gorm"
	"ru/kovardin/getapp/pkg/database"
)

// /conversion?client_id=1692468828256554387&yclid=12774450938537050111&install_timestamp=1692468885&appmetrica_device_id=3113866251430948486&click_id=&transaction_id=cpi14024587496244509100&match_type=fingerprint&tracker=appmetrica_172510023551860628

const (
	PartnerUnknown  = "unknown"
	PartnerVkads    = "vkads"
	PartnerYadirect = "yadirect"
)

func init() {
	database.AutoMigratedModels = append(database.AutoMigratedModels, Conversion{})
}

type Conversion struct {
	gorm.Model

	ClientId           string
	RbClickid          string
	Yclid              string
	InstallTimestamp   int
	AppMetricaDeviceId string
	TransactionId      string
	MatchType          string
	AppmetricaTracker  string
	ClickId            string
	Fire               bool
	Partner            string
	TrackerID          uint
	Tracker            Tracker
}
