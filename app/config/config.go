package config

import (
	"path"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/config"
	"go.uber.org/fx"

	admin "ru/kovardin/getapp/app/modules/admin/config"
	billing "ru/kovardin/getapp/app/modules/billing/config"
	boosty "ru/kovardin/getapp/app/modules/boosty/config"
	mediation "ru/kovardin/getapp/app/modules/mediation/config"
	http "ru/kovardin/getapp/app/servers/http/config"
	"ru/kovardin/getapp/pkg/cadence"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
	"ru/kovardin/getapp/pkg/mail"
)

type Application struct {
	fx.Out
	Logger   logger.Config
	Database database.Config
	Server   http.Config
	Mail     mail.Config
	Cadence  cadence.Config
}

type Modules struct {
	fx.Out
	Billing   billing.Config
	Boosty    boosty.Config
	Admin     admin.Config
	Mediation mediation.Config
}

type Config struct {
	fx.Out

	Application Application
	Modules     Modules
}

func New(env string, cfg string) Config {
	if env == "base" {
		panic("'base' can not be environment")
	}

	y, err := config.NewYAML(
		config.File(path.Join(cfg, "base.yml")),
		config.File(path.Join(cfg, env+".yml")),
	)
	if err != nil {
		panic(err)
	}

	c := Config{}
	err = y.Get("").Populate(&c)
	if err != nil {
		panic(err)
	}

	err = envconfig.Process(strings.ReplaceAll("lokalization", "-", ""), &c)
	if err != nil {
		panic(err)
	}

	return c
}
