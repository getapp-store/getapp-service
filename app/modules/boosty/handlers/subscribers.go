package handlers

import (
	"encoding/json"
	"kovardin.ru/projects/boosty/auth"
	"kovardin.ru/projects/boosty/request"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"kovardin.ru/projects/boosty"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/boosty/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Subscribers struct {
	log           *logger.Logger
	applications  *database.Repository[applications.Application]
	blogs         *database.Repository[models.Blog]
	subscribers   *database.Repository[models.Subscriber]
	subscriptions *database.Repository[models.Subscription]
}

func NewSubscribers(
	log *logger.Logger,
	applications *database.Repository[applications.Application],
	blogs *database.Repository[models.Blog],
	subscribers *database.Repository[models.Subscriber],
	subscriptions *database.Repository[models.Subscription],
) *Subscribers {
	return &Subscribers{
		log:           log,
		applications:  applications,
		blogs:         blogs,
		subscribers:   subscribers,
		subscriptions: subscriptions,
	}
}

type SubscriberResponse struct {
	Id           uint         `json:"id"`
	External     int          `json:"external"`
	Name         string       `json:"name"`
	Active       bool         `json:"active"`
	Amount       int          `json:"amount"`
	Subscription Subscription `json:"subscription"`
}

func (s *Subscribers) Subscriber(w http.ResponseWriter, r *http.Request) {
	// return subscriber subscriptions by application(blog)
	// refresh subscriptions on call
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		s.log.Error("error on parse application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := s.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		s.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	blog, err := s.blogs.First(database.Condition{
		In: map[string]any{
			"application_id": appid,
			"active":         true,
		},
	})
	if err != nil {
		s.log.Error("error on get blog from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if blog.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	external, err := strconv.Atoi(chi.URLParam(r, "external"))
	if err != nil {
		s.log.Error("error on parse external", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := auth.Info{}
	if err := json.Unmarshal([]byte(blog.Token), &token); err != nil {
		s.log.Error("error on parse boosty token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a, err := auth.New(
		auth.WithInfo(auth.Info{}),
		auth.WithInfoUpdateCallback(func(info auth.Info) {
			data, err := json.Marshal(info)
			if err != nil {
				s.log.Error("error on marshal data to info struct", zap.Error(err))
			}

			blog.Token = string(data)
			if err := s.blogs.Save(&blog); err != nil {
				s.log.Error("error on save boosty info struct to blog", zap.Error(err))
			}
		}),
	)

	rq, err := request.New(request.WithAuth(a))
	if err != nil {
		s.log.Error("error on prepare boosty lib request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := boosty.New(blog.Name, boosty.WithRequest(rq))
	if err != nil {
		s.log.Error("error on prepare boosty lib", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	current, err := b.Current()
	if err != nil {
		s.log.Error("error on load blog stats", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subscribers, err := b.Subscribers(0, current.PaidCount)
	if err != nil {
		s.log.Error("error on load blog subscribers", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subscriptions, err := b.Subscriptions(0, 20)
	if err != nil {
		s.log.Error("error on load blog subscriptions", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, subscriber := range subscribers {
		internalSubscriber, err := s.subscribers.First(database.Condition{
			In: map[string]any{
				"external": subscriber.ID,
			},
		})
		if err != nil {
			s.log.Error("error on load subscriber by external id", zap.Error(err), zap.Int("external", subscriber.ID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		internalSubscriber.External = subscriber.ID
		internalSubscriber.Name = subscriber.Name
		internalSubscriber.Email = subscriber.Email
		internalSubscriber.BlogID = blog.ID
		internalSubscriber.Amount = subscriber.Price * 100

		for _, subscription := range subscriptions {
			if subscription.ID == subscriber.Level.ID {
				// update subscription
				internalSubscription, err := s.subscriptions.First(database.Condition{
					In: map[string]any{
						"external": subscriber.Level.ID,
					},
					Joins: []string{
						"Blog",
					},
				})
				if err != nil {
					s.log.Error("error on load subscription by external id", zap.Int("external", subscriber.Level.ID), zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				internalSubscription.External = subscription.ID
				internalSubscription.Name = subscription.Name
				internalSubscription.Title = subscription.Name
				internalSubscription.BlogID = blog.ID
				internalSubscription.Amount = subscription.Price * 100
				internalSubscriber.SubscriptionID = internalSubscription.ID

				if internalSubscription.ID == 0 {
					// create new
					internalSubscription.Active = !(subscription.Deleted || subscription.IsArchived)

					if err := s.subscriptions.Create(&internalSubscription); err != nil {
						s.log.Error("error on create subscription in db", zap.Int("external", subscription.ID), zap.Error(err))
						continue
					}
				} else {
					// update exists
					if err := s.subscriptions.Save(&internalSubscription); err != nil {
						s.log.Error("error on save subscription in db", zap.Int("external", subscription.ID), zap.Error(err))
						continue
					}
				}
				internalSubscriber.Subscription = internalSubscription
				break
			}
		}

		if internalSubscriber.ID == 0 {
			// create new
			internalSubscriber.Active = subscriber.Subscribed

			if err := s.subscribers.Create(&internalSubscriber); err != nil {
				s.log.Error("error on create subscriber in db", zap.Int("external", subscriber.ID), zap.Error(err))
				continue
			}
		} else {
			// update exists
			if err := s.subscribers.Save(&internalSubscriber); err != nil {
				s.log.Error("error on save subscriber in db", zap.Int("external", subscriber.ID), zap.Error(err))
				continue
			}
		}

		if subscriber.ID == external {
			render.JSON(w, r, SubscriberResponse{
				Id:       internalSubscriber.ID,
				External: internalSubscriber.External,
				Name:     internalSubscriber.Name,
				Amount:   internalSubscriber.Amount,
				Active:   internalSubscriber.Active,
				Subscription: Subscription{
					Id:       internalSubscriber.Subscription.ID,
					External: internalSubscriber.Subscription.External,
					Blog:     internalSubscriber.Subscription.Blog.Name,
					Name:     internalSubscriber.Subscription.Name,
					Title:    internalSubscriber.Subscription.Title,
					Amount:   internalSubscriber.Subscription.Amount,
					Active:   internalSubscriber.Subscription.Active,
				},
			})

			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
