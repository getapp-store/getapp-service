package main

import (
	"github.com/qor5/admin/presets/gorm2op"
	"math/rand"
	"os"
	"time"

	"github.com/qor5/admin/presets"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/config"
	"ru/kovardin/getapp/app/modules"
	_ "ru/kovardin/getapp/app/modules/admin"
	_ "ru/kovardin/getapp/app/modules/ads"
	_ "ru/kovardin/getapp/app/modules/applications"
	_ "ru/kovardin/getapp/app/modules/billing"
	_ "ru/kovardin/getapp/app/modules/boosty"
	_ "ru/kovardin/getapp/app/modules/deployer"
	_ "ru/kovardin/getapp/app/modules/landings"
	_ "ru/kovardin/getapp/app/modules/lokalize"
	_ "ru/kovardin/getapp/app/modules/mediation"
	_ "ru/kovardin/getapp/app/modules/tracker"
	_ "ru/kovardin/getapp/app/modules/users"
	_ "ru/kovardin/getapp/app/modules/warehouse"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/cadence"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
	"ru/kovardin/getapp/pkg/mail"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	cmds := cli.Commands{
		&cli.Command{
			Name: "server",
			Action: func(c *cli.Context) error {
				setup(c, append(
					modules.Invokes,
					fx.Invoke(func(server *http.Server) {}),
					fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: log}
					}),
				)...).Run()

				return nil
			},
		},
	}

	for _, c := range modules.Commands {
		cmds = append(cmds, c(setup))
	}

	app := &cli.App{
		Name:  "getapp",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env",
				Value: "dev",
				Usage: "environment",
			},
			&cli.StringFlag{
				Name:  "configs",
				Value: "./configs",
				Usage: "configs path",
			},
		},
		Commands: cmds,
	}

	app.Run(os.Args)
}

func setup(c *cli.Context, opts ...fx.Option) *fx.App {

	env := c.String("env")
	cfg := c.String("configs")

	opts = append(opts, fx.Provide(
		func() config.Config {
			return config.New(env, cfg)
		},
		func(db *database.Database) *presets.Builder {
			return presets.New().
				DataOperator(gorm2op.DataOperator(db.DB()))
		},
		http.New,
		logger.New,
		database.NewDatabase,
		mail.New,
		cadence.New,
	))
	opts = append(opts, modules.Providers...)
	opts = append(opts, fx.StartTimeout(time.Second*60))
	opts = append(opts, fx.NopLogger)

	return fx.New(
		opts...,
	)
}
