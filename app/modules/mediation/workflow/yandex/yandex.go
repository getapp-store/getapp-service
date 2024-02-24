package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/app/modules/mediation/networks"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

const layout = "2006-01-02"

type Yandex struct {
	log         *logger.Logger
	ecpm        *database.Repository[models.Cpm]
	impressions *database.Repository[models.Impression]
	units       *database.Repository[models.Unit]
	client      *http.Client
	url         string
}

func New(
	log *logger.Logger,
	ecpm *database.Repository[models.Cpm],
	impressions *database.Repository[models.Impression],
	units *database.Repository[models.Unit],
) *Yandex {
	return &Yandex{
		log:         log,
		ecpm:        ecpm,
		impressions: impressions,
		units:       units,
		client:      &http.Client{},
		url:         `https://partner2.yandex.ru/api/statistics2/get.json`,
	}
}

func (y *Yandex) Execute(ctx context.Context, name string) (string, error) {
	uu, err := y.units.Find(database.Condition{
		In: map[string]any{
			`"units"."active"`: true,
			`"Network"."name"`: networks.Yandex,
		},
		Joins: []string{
			"Network",
		},
	})
	if err != nil {
		y.log.Error("error on getting trackers", zap.Error(err))
	}

	for _, u := range uu {
		if err := y.process(u); err != nil {
			y.log.Error("error parse cpm", zap.Error(err))
		}
	}

	y.log.Info("finished yandex ecpm activity")

	return "parsed yandex ecpm", nil
}

func (y *Yandex) process(model models.Unit) error {
	to := time.Now()
	currency := "USD"
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	u := y.url + "?"
	u += `currency=` + currency + `&stat_type=main&lang=ru&dimension_field=date|day&entity_field=complex_block_id&field=impressions&field=partner_wo_nds&field=ecpm_partner_wo_nds&`
	u += `&period=` + to.Format(layout) + `&period=` + to.Format(layout) + `&`
	u += `limits=` + url.QueryEscape(`{"limit":200,"offset":0}`) + `&`
	u += `order_by=[` + url.QueryEscape(`{"field":"date","dir":"asc"}`) + `]&`
	u += `filter=[` + url.QueryEscape(`"complex_block_id", "=", "`+model.Unit+`"`) + `]`

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "OAuth "+model.Data)

	resp, err := y.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("yandex partner api status %d", resp.StatusCode)
	}

	data := Stats{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Data.Points) == 0 {
		return nil
	}

	if len(data.Data.Points[0].Measures) == 0 {
		return nil
	}

	cmp := data.Data.Points[0].Measures[0].EcpmPartnerWoNds

	if err := y.ecpm.Save(&models.Cpm{
		UnitId:      model.ID,
		NetworkId:   model.NetworkId,
		PlacementId: model.PlacementId,
		Date:        time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, time.UTC),
		Amount:      cmp,
		CreatedAt:   to,
	}); err != nil {
		return err
	}

	return nil
}
