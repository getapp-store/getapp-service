package handlers

import (
	"net/http"

	"encoding/json"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/bidding"
	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Auction struct {
	log        *logger.Logger
	bidding    *bidding.Bidding
	placements *database.Repository[models.Placement]
	cpms       *database.Repository[models.Cpm]
	units      *database.Repository[models.Unit]
}

func NewAuction(
	log *logger.Logger,
	bidding *bidding.Bidding,
	placements *database.Repository[models.Placement],
	cpms *database.Repository[models.Cpm],
	units *database.Repository[models.Unit],
) *Auction {
	return &Auction{
		log:        log,
		bidding:    bidding,
		placements: placements,
		cpms:       cpms,
		units:      units,
	}
}

type User struct {
	Id string `json:"id"`
}

type BidRequest struct {
	Unit string `json:"unit"`
	User User   `json:"user"`
}

type BidResponse struct {
	Unit string  `json:"unit"`
	Cpm  float64 `json:"cpm"`
	Bid  float64 `json:"bid"`
}

func (a *Auction) Bid(w http.ResponseWriter, r *http.Request) {
	req := BidRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.log.Error("error on parse request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	unit, err := a.units.First(database.Condition{
		In: map[string]any{
			"unit":   req.Unit,
			"active": true,
		},
	})
	if err != nil {
		a.log.Error("error on get unit", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// расчет цены показа
	cpm, err := a.bidding.Cpm(unit)
	if err != nil {
		a.log.Error("error on find cpm", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// расчет ставки
	bid, err := a.bidding.Bid(unit)
	if err != nil {
		a.log.Error("error on find cpm", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// расчет ставки с учетом теста
	// тут bid может увеличится
	bid, err = a.bidding.Bandit(bid, unit)
	if err != nil {
		a.log.Error("error on find cpm", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := BidResponse{
		Unit: req.Unit,
		Bid:  bid,
		Cpm:  cpm,
	}

	render.JSON(w, r, resp)
}
