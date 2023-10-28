package auth

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	users "ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/utils"
)

const ApplicationIdKey = "ApplicationIdKey"
const UserIdKey = "UserIdKey"

type Auth struct {
	logger       *zap.Logger
	applications *database.Repository[applications.Application]
	users        *database.Repository[users.User]
}

func New(logger *zap.Logger, applications *database.Repository[applications.Application], users *database.Repository[users.User]) *Auth {
	return &Auth{
		logger:       logger,
		applications: applications,
		users:        users,
	}
}

func (a *Auth) UserAuthorize() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userToken := utils.TokenFromHeader(r, utils.UserKey)

			if userToken == "" {
				userToken = utils.TokenFromParams(r, utils.UserKey)
			}

			if userToken == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := a.users.First(database.Condition{
				In: map[string]any{
					"api_token": userToken,
				},
			})

			if err != nil {
				a.logger.Error("error on find user by userToken", zap.Error(err), zap.String("appToken", userToken))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if user.ID == 0 {
				a.logger.Error("user not found", zap.String("appToken", userToken))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), UserIdKey, user.ID))

			next.ServeHTTP(w, r)
		})
	}
}
func (a *Auth) AppAuthorize() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appToken := utils.TokenFromHeader(r, utils.AppKey)

			if appToken == "" {
				appToken = utils.TokenFromParams(r, utils.AppKey)
			}

			if appToken == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// get application by appToken
			// set app id to request context
			app, err := a.applications.First(database.Condition{
				In: map[string]any{
					"api_token": appToken,
				},
			})

			// get user appToken

			if err != nil {
				a.logger.Error("error on find application by appToken", zap.Error(err), zap.String("appToken", appToken))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if app.ID == 0 {
				a.logger.Error("app not found", zap.String("appToken", appToken))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), ApplicationIdKey, app.ID))

			next.ServeHTTP(w, r)
		})
	}
}
