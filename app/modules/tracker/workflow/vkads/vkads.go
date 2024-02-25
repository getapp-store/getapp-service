package vkads

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Vkads struct {
	log         *logger.Logger
	conversions *database.Repository[models.Conversion]
	trackers    *database.Repository[models.Tracker]
	client      *http.Client
}

func New(
	log *logger.Logger,
	conversions *database.Repository[models.Conversion],
	trackers *database.Repository[models.Tracker],
) *Vkads {
	return &Vkads{
		log:         log,
		conversions: conversions,
		trackers:    trackers,
		client:      &http.Client{},
	}
}

func (u *Vkads) Execute(ctx context.Context, name string) (string, error) {
	tt, err := u.trackers.Find(database.Condition{
		In: map[string]any{
			"active": true,
		},
	})
	if err != nil {
		u.log.Error("error on getting trackers", zap.Error(err))
	}

	for _, t := range tt {
		u.process(t)
	}

	return "uploaded vk tracker data", nil
}

func (u *Vkads) process(tracker models.Tracker) {
	conversions, err := u.conversions.Find(database.Condition{
		In: map[string]any{
			"fire":       false,
			"partner":    models.PartnerVkads,
			"tracker_id": tracker.ID,
		},
	})
	if err != nil {
		u.log.Error("error on get vkads conversions", zap.Error(err))
		return
	}

	if len(conversions) == 0 {
		return
	}

	if err := u.vkads(conversions, tracker.VkTracker); err != nil {
		u.log.Error("error on upload conversions to vkads", zap.Error(err))
		return
	}

	u.fire(conversions)
}

func (u *Vkads) vkads(items []models.Conversion, vkurl string) error {
	for _, item := range items {
		link := vkurl + item.RbClickid

		u.log.Warn("vkads link for check", zap.String("link", link))

		if _, err := u.client.Get(link); err != nil {
			u.log.Error("error on send vk pixel", zap.Error(err))
			continue
		}

	}
	return nil
}

func (u *Vkads) fire(items []models.Conversion) {
	for i := range items {
		func(item models.Conversion) {
			item.Fire = true
			if err := u.conversions.Save(&item); err != nil {
				u.log.Error("error on fire conversion", zap.Error(err))
			}
		}(items[i])
	}
}
