package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
	"kovardin.ru/projects/boosty"
	"kovardin.ru/projects/boosty/auth"
	"kovardin.ru/projects/boosty/request"

	"ru/kovardin/getapp/app/modules/boosty/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Parser struct {
	log           *logger.Logger
	blogs         *database.Repository[models.Blog]
	subscriptions *database.Repository[models.Subscription]
	subscribers   *database.Repository[models.Subscriber]
	client        *http.Client
}

func New(
	log *logger.Logger,
	blogs *database.Repository[models.Blog],
	subscriptions *database.Repository[models.Subscription],
	subscribers *database.Repository[models.Subscriber],
) *Parser {
	return &Parser{
		log:           log,
		blogs:         blogs,
		subscriptions: subscriptions,
		subscribers:   subscribers,
		client:        &http.Client{},
	}
}

func (p *Parser) Execute(ctx context.Context, name string) (string, error) {
	bb, err := p.blogs.Find(database.Condition{
		In: map[string]any{
			"active": true,
		},
	})
	if err != nil {
		p.log.Error("error on getting blogs", zap.Error(err))
	}

	errs := []error{}
	for _, b := range bb {
		if err := p.process(b); err != nil {
			errs = append(errs, err)
		}
	}

	p.log.Info("finished boosty parser activity")

	return fmt.Sprintf("parsed boosty: %+v", errs), nil
}

func (p *Parser) process(b models.Blog) error {
	p.log.Info("parse blog", zap.String("blog", b.Url), zap.String("token", b.Token))

	token := auth.Info{}
	if err := json.Unmarshal([]byte(b.Token), &token); err != nil {
		return fmt.Errorf("error on parse boosty token: %w", err)
	}

	a, err := auth.New(
		auth.WithInfo(token),
		auth.WithInfoUpdateCallback(func(info auth.Info) {
			data, err := json.Marshal(info)
			if err != nil {
				p.log.Error("error on marshal data to info struct", zap.Error(err))
			}

			b.Token = string(data)
			if err := p.blogs.Save(&b); err != nil {
				p.log.Error("error on save boosty info struct to blog", zap.Error(err))
			}
		}),
	)
	if err != nil {
		return fmt.Errorf("error on prepare boosty lib auth: %w", err)
	}

	rq, err := request.New(request.WithAuth(a))
	if err != nil {
		return fmt.Errorf("error on prepare boosty lib request: %w", err)
	}

	api, err := boosty.New(b.Name, boosty.WithRequest(rq))
	if err != nil {
		return fmt.Errorf("error on prepare boosty lib: %w", err)
	}

	v := url.Values{}
	v.Add("offset", "0")
	v.Add("limit", "20")
	v.Add("order", "gt")
	v.Add("sort_by", "on_time")

	// load subscriptions
	subscriptions, err := api.Subscriptions(v)
	if err != nil {
		return fmt.Errorf("error on fetch subscriptions blog %s: %w", b.Name, err)
	}

	// save to db
	for _, s := range subscriptions.Data {
		// check if exist

		model, err := p.subscriptions.First(database.Condition{
			In: map[string]any{
				"external": s.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("error on get subscription from db external %d: %w", s.ID, err)
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
				return fmt.Errorf("error on create subscription in db external %d: %w", s.ID, err)
			}
		} else {
			// update exists
			if err := p.subscriptions.Save(&model); err != nil {
				return fmt.Errorf("error on save subscription in db external %d: %w", s.ID, err)
			}
		}
	}

	current, err := api.Current()
	if err != nil {
		return fmt.Errorf("error on get stats blog %s: %w", b.Name, err)
	}

	v = url.Values{}
	v.Add("offset", "0")
	v.Add("limit", fmt.Sprintf("%d", current.FollowersCount+current.PaidCount))
	v.Add("order", "gt")
	v.Add("sort_by", "on_time")

	// fetch subscribers
	subscribers, err := api.Subscribers(v)
	if err != nil {
		return fmt.Errorf("error on fetch subscribers blog %s: %w", b.Name, err)
	}

	// save to db
	for _, s := range subscribers.Data {
		model, err := p.subscribers.First(database.Condition{
			In: map[string]any{
				"external": s.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("error on get subscriber from db external %d: %w", s.ID, err)
		}

		subscription, err := p.subscriptions.First(database.Condition{
			In: map[string]any{
				"external": s.Level.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("error on get subscription by subscriber from db external %d: %w", s.ID, err)
		}

		model.External = s.ID
		model.Name = s.Name
		model.Email = s.Email
		model.BlogID = b.ID
		model.SubscriptionID = subscription.ID
		model.Amount = s.Price * 100
		model.Active = subscription.Active // len(s.Level.Data) > 0 - у бесплатной подписки поле Data без данных

		if model.ID == 0 {
			// create new
			model.Active = s.Subscribed

			if err := p.subscribers.Create(&model); err != nil {
				return fmt.Errorf("error on create subscriber in db external %d: %w", s.ID, err)
			}
		} else {
			// update exists
			if err := p.subscribers.Save(&model); err != nil {
				return fmt.Errorf("error on save subscriber in db external %d: %w", s.ID, err)
			}
		}
	}

	return nil
}
