package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Placements struct {
	log        *logger.Logger
	placements *database.Repository[models.Placement]
}

func NewPlacements(log *logger.Logger, networks *database.Repository[models.Placement]) *Placements {
	return &Placements{
		log:        log,
		placements: networks,
	}
}

type Unit struct {
	Name      string `json:"name"`
	Unit      string `json:"unit"`
	Network   string `json:"network"`
	Placement int    `json:"placement"`
}

type PlacementResponse struct {
	Name      string `json:"name"`
	Placement int    `json:"placement"`
	Format    string `json:"format"`
	Units     []Unit `json:"units"`
}

func (h *Placements) Placement(w http.ResponseWriter, r *http.Request) {
	placement, _ := strconv.Atoi(chi.URLParam(r, "placement"))

	p, err := h.placements.First(database.Condition{
		In: map[string]any{
			"id":     placement,
			"active": true,
		},
		Preload: []string{"Units", "Units.Network"},
	})
	if err != nil {
		h.log.Error("error on load placement", zap.Error(err), zap.Int("placement", placement))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := PlacementResponse{
		Placement: placement,
		Name:      p.Name,
		Format:    p.Format,
		Units:     []Unit{},
	}

	for _, u := range p.Units {
		if u.Active == false {
			continue
		}

		if u.Network.Active == false {
			continue
		}

		resp.Units = append(resp.Units, Unit{
			Name:      u.Name,
			Unit:      u.Unit,
			Network:   u.Network.Name,
			Placement: placement,
		})
	}

	render.JSON(w, r, resp)
}
