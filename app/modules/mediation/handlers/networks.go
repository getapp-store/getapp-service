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

type Networks struct {
	log      *logger.Logger
	networks *database.Repository[models.Network]
}

func NewNetworks(log *logger.Logger, networks *database.Repository[models.Network]) *Networks {
	return &Networks{
		log:      log,
		networks: networks,
	}
}

type Network struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type NetworksResponse struct {
	Id       int       `json:"id"` // application id
	Networks []Network `json:"networks"`
}

func (h *Networks) Networks(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		h.log.Error("error on parse application", zap.Error(err), zap.Int("app", id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get networks by appid

	nn, err := h.networks.Find(database.Condition{
		In: map[string]any{
			"active":         true,
			"application_id": id,
		},
	})

	if err != nil {
		h.log.Error("error on get networks", zap.Error(err), zap.Int("app", id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := NetworksResponse{
		Id:       id,
		Networks: []Network{},
	}

	for _, n := range nn {
		resp.Networks = append(resp.Networks, Network{
			Id:   n.ID,
			Name: n.Name,
			Key:  n.Key,
		})
	}

	render.JSON(w, r, resp)
}
