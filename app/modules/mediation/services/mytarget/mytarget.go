package mytarget

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ru/kovardin/getapp/app/modules/mediation/networks"
	"strconv"
	"time"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

const layout = "2006-01-02"

type Stats struct {
	Items []Item `json:"items"`
	Total struct {
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"total"`
}

type Item struct {
	Rows []struct {
		Date      string `json:"date"`
		Currency  string `json:"currency"`
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"rows"`
	Total struct {
		Histogram []Data `json:"histogram"`
		Total     Data   `json:"total"`
	} `json:"total"`
	ID int `json:"id"`
}

type Data struct {
	Clicks           int     `json:"clicks"`
	Shows            int     `json:"shows"`
	Goals            int     `json:"goals"`
	Noshows          int     `json:"noshows"`
	Requests         int     `json:"requests"`
	RequestedBanners int     `json:"requested_banners"`
	ResponsedBlocks  int     `json:"responsed_blocks"`
	ResponsedBanners int     `json:"responsed_banners"`
	Amount           string  `json:"amount"`
	Responses        int     `json:"responses"`
	Cpm              string  `json:"cpm"`
	Ctr              int     `json:"ctr"`
	FillRate         float64 `json:"fill_rate"`
	ShowRate         float64 `json:"show_rate"`
	Vtr              int     `json:"vtr"`
	Vr               int     `json:"vr"`
}

type MyTarget struct {
	log    *logger.Logger
	client *http.Client
	period time.Duration
	units  *database.Repository[models.Unit]
	ecpms  *database.Repository[models.Cpm]
	url    string
}

func New(log *logger.Logger, units *database.Repository[models.Unit], ecpms *database.Repository[models.Cpm]) *MyTarget {
	return &MyTarget{
		log:    log,
		units:  units,
		ecpms:  ecpms,
		client: &http.Client{},
		period: time.Hour * 1,
		//period: time.Second * 10,

		url: "https://target.my.com/api/v2/statistics/geo/pads/hour.json?id=%s&date_from=%s&date_to=%s",
	}
}

func (m *MyTarget) Start() {
	// get from api
	// pars response
	// count cpm and save to db

	ticker := time.NewTicker(m.period)
	go func() {
		for ; true; <-ticker.C {
			uu, err := m.units.Find(database.Condition{
				In: map[string]any{
					`"units"."active"`: true,
					`"Network"."name"`: networks.MyTarget,
				},
				Joins: []string{
					"Network",
				},
			})
			if err != nil {
				m.log.Error("error on getting trackers", zap.Error(err))
			}

			for _, u := range uu {
				m.process(u)
			}
		}

		//for range ticker.C {
		//	u.process()
		//}
	}()
}

func (m *MyTarget) process(model models.Unit) {
	m.log.Info("unit", zap.Any("unit", model))

	now := time.Now()

	from := now.Add(-time.Hour * 24 * 2).Format(layout)
	to := now.Format(layout)

	url := fmt.Sprintf(m.url, model.Data, from, to)

	m.log.Info("final url", zap.String("url", url))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		m.log.Error("error on create request", zap.Error(err))

		return
	}

	req.Header.Set("Authorization", "Bearer "+model.Network.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		m.log.Error("error on make request", zap.Error(err))

		return
	}

	if resp.StatusCode != http.StatusOK {
		m.log.Error("error response code", zap.Int("code", resp.StatusCode))

		return
	}

	stats := &Stats{}

	if err := json.NewDecoder(resp.Body).Decode(stats); err != nil {
		m.log.Error("error pase response", zap.Error(err))

		return
	}

	cpm, err := strconv.ParseFloat(stats.Total.Total.Cpm, 64)
	if err != nil {
		m.log.Error("error pase stats", zap.Error(err))

		return
	}

	// из mytracker все приходит в рублях, перевести cpm нужно в доллары
	converted := cpm * networks.USD

	if err := m.ecpms.Save(&models.Cpm{
		UnitId:      model.ID,
		NetworkId:   model.NetworkId,
		PlacementId: model.PlacementId,
		Date:        time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC),
		Amount:      converted, // переводим в копейки
		CreatedAt:   now,
	}); err != nil {
		m.log.Error("error save cpm", zap.Error(err))
	}
}

func (m *MyTarget) Stop() {

}
