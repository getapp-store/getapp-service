package bigo

import (
	"ru/kovardin/getapp/app/modules/mediation/networks"
	"time"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Bigo struct {
	log         *logger.Logger
	ecpm        *database.Repository[models.Cpm]
	impressions *database.Repository[models.Impression]
	units       *database.Repository[models.Unit]

	period time.Duration
}

func New(
	log *logger.Logger,
	ecpm *database.Repository[models.Cpm],
	impressions *database.Repository[models.Impression],
	units *database.Repository[models.Unit],
) *Bigo {
	return &Bigo{
		log:         log,
		ecpm:        ecpm,
		impressions: impressions,
		units:       units,

		period: time.Hour * 1,
		//period: time.Second * 10,
	}
}

func (b *Bigo) Start() {
	ticker := time.NewTicker(b.period)
	go func() {
		for ; true; <-ticker.C {
			uu, err := b.units.Find(database.Condition{
				In: map[string]any{
					`"units"."active"`: true,
					`"Network"."name"`: networks.Bigo,
				},
				Joins: []string{
					"Network",
				},
			})
			if err != nil {
				b.log.Error("error on getting trackers", zap.Error(err))
			}

			for _, u := range uu {
				b.process(u)
			}
		}

		//for range ticker.C {
		//	u.process()
		//}
	}()
}

func (y *Bigo) process(model models.Unit) error {
	// собираем показы за период и считаем CPM
	to := time.Now()
	from := to.Add(-time.Hour * 24 * 3)
	ii, err := y.impressions.Find(database.Condition{
		In: map[string]any{
			"unit_id":      model.ID,
			"network_id":   model.NetworkId,
			"placement_id": model.PlacementId,
		},
		Where: []database.Where{
			{
				Condition: "date > ? and date <= ?",
				Values:    []any{from, to},
			},
		},
	})

	if err != nil {
		return err
	}

	var (
		total, cmp float64
		cnt        int
	)

	for _, i := range ii {
		y.log.Info("yandex impressions service", zap.Any("impression", i))

		total += i.Revenue
		cnt++
	}

	if cnt != 0 {
		// средняя цена одного показа за период
		one := total / float64(cnt)

		// цена тысячи показов
		cmp = one * 1000
	}

	if err := y.ecpm.Save(&models.Cpm{
		UnitId:      model.ID,
		NetworkId:   model.NetworkId,
		PlacementId: model.PlacementId,
		Date:        time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, time.UTC),
		Amount:      cmp, // переводим в копей
		CreatedAt:   to,
	}); err != nil {
		y.log.Error("error save cpm", zap.Error(err))
	}

	return nil
}

func (b *Bigo) Stop() {

}
