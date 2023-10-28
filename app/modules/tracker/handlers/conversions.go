package handlers

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Conversions struct {
	log         *logger.Logger
	conversions *database.Repository[models.Conversion]
}

func NewConversions(log *logger.Logger, conversions *database.Repository[models.Conversion]) *Conversions {
	return &Conversions{
		log:         log,
		conversions: conversions,
	}
}

func (t *Conversions) Fire(w http.ResponseWriter, r *http.Request) {
	// https://service.getapp.store/v1/fire/?client_id=1692468828256554387&yclid=12774450938537050111&install_timestamp=1692468885&appmetrica_device_id=3113866251430948486&click_id=&transaction_id=cpi14024587496244509100&match_type=fingerprint&tracker=appmetrica_172510023551860628
	// save data to database
	install, err := strconv.Atoi(r.URL.Query().Get("install_timestamp"))
	if err != nil {
		t.log.Error("error on get install_timestamp", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	partner := r.URL.Query().Get("partner")
	yclid := r.URL.Query().Get("yclid")
	client := r.URL.Query().Get("client_id")
	rbclickid := r.URL.Query().Get("rb_clickid")

	if yclid == "" && client == "" && rbclickid == "" {
		t.log.Error("all params are empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if yclid != "" {
		partner = models.PartnerYadirect
	} else if rbclickid != "" {
		partner = models.PartnerVkads
	} else if partner != "" {
		partner = partner
	}

	if partner != models.PartnerYadirect && partner != models.PartnerVkads {
		t.log.Error("unknown partner", zap.String("partner", partner))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := t.conversions.Create(&models.Conversion{
		ClientId:           client,
		Yclid:              yclid,
		InstallTimestamp:   install,
		AppMetricaDeviceId: r.URL.Query().Get("appmetrica_device_id"),
		TransactionId:      r.URL.Query().Get("transaction_id"),
		MatchType:          r.URL.Query().Get("match_type"),
		AppmetricaTracker:  r.URL.Query().Get("tracker"),
		ClickId:            r.URL.Query().Get("click_id"),
		Fire:               false,
		Partner:            partner,
		RbClickid:          rbclickid,
		TrackerID:          1,
	}); err != nil {
		t.log.Error("error on save conversion", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("^)"))
}
