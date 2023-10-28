package services

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"kovardin.ru/projects/boosty"
	"ru/kovardin/getapp/app/modules/boosty/models"
	"ru/kovardin/getapp/pkg/database"
)

type Parser struct {
	log           *zap.Logger
	blogs         *database.Repository[models.Blog]
	subscriptions *database.Repository[models.Subscription]
	subscribers   *database.Repository[models.Subscriber]
	client        *http.Client
	period        time.Duration
}

func NewParser(log *zap.Logger, blogs *database.Repository[models.Blog], subscriptions *database.Repository[models.Subscription], subscribers *database.Repository[models.Subscriber]) *Parser {
	return &Parser{
		log:           log,
		blogs:         blogs,
		subscriptions: subscriptions,
		subscribers:   subscribers,
		client:        &http.Client{},
		period:        time.Hour * 1,
	}
}

func (p *Parser) Start() {
	ticker := time.NewTicker(p.period)
	go func() {
		for ; true; <-ticker.C {
			bb, err := p.blogs.Find(database.Condition{
				In: map[string]any{
					"active": true,
				},
			})
			if err != nil {
				p.log.Error("error on getting blogs", zap.Error(err))
			}

			for _, b := range bb {
				p.process(b)
			}
		}
	}()
}

func (p *Parser) Stop() {

}

func (p *Parser) process(b models.Blog) {
	p.log.Info("parse blog", zap.String("blog", b.Url), zap.String("token", b.Token))

	api := boosty.New(b.Name, b.Token)

	// load subscriptions
	subscriptions, err := api.Subscriptions(0, 20)
	if err != nil {
		p.log.Error("error on fetch subscriptions", zap.String("blog", b.Name), zap.Error(err))
		return
	}

	// save to db
	for _, s := range subscriptions {
		// check if exist

		model, err := p.subscriptions.First(database.Condition{
			In: map[string]any{
				"external": s.ID,
			},
		})
		if err != nil {
			p.log.Error("error on get subscription from db", zap.Int("external", s.ID), zap.Error(err))
			continue
		}

		model.External = s.ID
		model.Name = s.Name
		model.Title = s.Name
		model.BlogID = b.ID
		model.Amount = s.Price * 100

		if model.ID == 0 {
			// create new
			model.Active = !(s.Deleted || s.IsArchived)

			if err := p.subscriptions.Create(&model); err != nil {
				p.log.Error("error on create subscription in db", zap.Int("external", s.ID), zap.Error(err))
				continue
			}
		} else {
			// update exists
			if err := p.subscriptions.Save(&model); err != nil {
				p.log.Error("error on save subscription in db", zap.Int("external", s.ID), zap.Error(err))
				continue
			}
		}
	}

	stats, err := api.Stats()
	if err != nil {
		p.log.Error("error on get stats", zap.String("blog", b.Name), zap.Error(err))
		return
	}

	// fetch subscribers
	subscribers, err := api.Subscribers(0, stats.FollowersCount+stats.PaidCount)
	if err != nil {
		p.log.Error("error on fetch subscribers", zap.String("blog", b.Name), zap.Error(err))
		return
	}

	// save to db
	for _, s := range subscribers {
		model, err := p.subscribers.First(database.Condition{
			In: map[string]any{
				"external": s.ID,
			},
		})
		if err != nil {
			p.log.Error("error on get subscriber from db", zap.Int("external", s.ID), zap.Error(err))
			continue
		}

		subscription, err := p.subscriptions.First(database.Condition{
			In: map[string]any{
				"external": s.Level.ID,
			},
		})
		if err != nil {
			p.log.Error("error on get subscription by subscriber from db", zap.Int("external", s.ID), zap.Error(err))
			continue
		}

		model.External = s.ID
		model.Name = s.Name
		model.Email = s.Email
		model.BlogID = b.ID
		model.SubscriptionID = subscription.ID
		model.Amount = s.Price * 100

		if model.ID == 0 {
			// create new
			model.Active = s.Subscribed

			if err := p.subscribers.Create(&model); err != nil {
				p.log.Error("error on create subscriber in db", zap.Int("external", s.ID), zap.Error(err))
				continue
			}
		} else {
			// update exists
			if err := p.subscribers.Save(&model); err != nil {
				p.log.Error("error on save subscriber in db", zap.Int("external", s.ID), zap.Error(err))
				continue
			}
		}
	}
}
