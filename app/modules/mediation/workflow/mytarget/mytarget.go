package mytarget

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

const layout = "2006-01-02"

type MyTarget struct {
	log    *logger.Logger
	client *http.Client
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
		url:    "https://target.my.com/api/v2/statistics/geo/pads/hour.json?id=%s&date_from=%s&date_to=%s",
	}
}

func (m *MyTarget) Execute(ctx context.Context, name string) (string, error) {
	uu, err := m.units.Find(database.Condition{
		In: map[string]any{
			`"units"."active"`: true,
			`"Network"."name"`: models.MyTargetNetwork,
		},
		Joins: []string{
			"Network",
		},
	})
	if err != nil {
		m.log.Error("error on getting trackers", zap.Error(err))
	}

	for _, u := range uu {
		if err := m.process(u); err != nil {
			m.log.Error("error parse cpm", zap.Error(err))
		}
	}

	m.log.Info("finished mytarget ecpm activity")

	return "parsed mytarget ecpm", nil
}

func (m *MyTarget) process(model models.Unit) error {
	m.log.Info("unit", zap.Any("unit", model))

	now := time.Now()

	from := now.Add(-time.Hour * 24 * 2).Format(layout)
	to := now.Format(layout)

	url := fmt.Sprintf(m.url, model.Data, from, to)

	m.log.Info("final url", zap.String("url", url))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+model.Network.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	stats := &Stats{}

	if err := json.NewDecoder(resp.Body).Decode(stats); err != nil {
		return err
	}

	cpm, err := strconv.ParseFloat(stats.Total.Total.Cpm, 64)
	if err != nil {
		return err
	}

	// из mytracker все приходит в рублях, перевести cpm нужно в доллары
	converted := cpm * models.USD

	return m.ecpms.Save(&models.Cpm{
		UnitId:      model.ID,
		NetworkId:   model.NetworkId,
		PlacementId: model.PlacementId,
		Date:        time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC),
		Amount:      converted, // переводим в копейки
		CreatedAt:   now,
	})
}
