package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"strconv"
	"time"

	"ru/kovardin/getapp/pkg/logger"
)

// save impression revenue
// save impression fact

type Impressions struct {
	log         *logger.Logger
	impressions *database.Repository[models.Impression]
	units       *database.Repository[models.Unit]
}

func NewImpressions(
	log *logger.Logger,
	impressions *database.Repository[models.Impression],
	units *database.Repository[models.Unit],
) *Impressions {
	return &Impressions{
		log:         log,
		impressions: impressions,
		units:       units,
	}
}

type ImpressionRequest struct {
	Unit    string  `json:"unit"`
	Data    string  `json:"data"`
	Revenue float64 `json:"revenue"`
}

func (i *Impressions) Impression(w http.ResponseWriter, r *http.Request) {
	// {"data":"{"currency":"RUB","revenueUSD":"0.000835340","precision":"estimated","revenue":"0.073805629","requestId":"1705527666378952-53598225983379352300275-production-app-host-sas-pcode-249","blockId":"R-M-2768512-2","adType":"interstitial","ad_unit_id":"R-M-2768512-2","network":{"name":"Yandex","adapter":"Yandex","ad_unit_id":"R-M-2768512-2"}}","price":150.0,"unit":"R-M-2768512-2"}
	placement, _ := strconv.Atoi(chi.URLParam(r, "placement"))
	req := ImpressionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		i.log.Error("error on parse request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	unit, err := i.units.First(database.Condition{
		In: map[string]any{
			"unit":         req.Unit,
			"placement_id": placement,
			"active":       true,
		},
		Preload: []string{
			"Network",
		},
	})
	if err != nil {
		i.log.Error("error get unit", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err := i.impressions.Create(&models.Impression{
		UnitId:      unit.ID,
		NetworkId:   unit.NetworkId,
		PlacementId: uint(placement),
		Revenue:     req.Revenue,
		Date:        time.Now(),
		Raw:         req.Data,
	}); err != nil {
		i.log.Error("error save impression", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
