package mediation

import (
	"context"
	"fmt"
	"log"
	"os"
	"ru/kovardin/getapp/app/modules/mediation/rotation/bidding"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/mediation/config"
	"ru/kovardin/getapp/app/modules/mediation/handlers"
	"ru/kovardin/getapp/app/modules/mediation/models"
	"ru/kovardin/getapp/app/modules/mediation/repos"
	"ru/kovardin/getapp/app/modules/mediation/services/parser"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/bigo"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/cpa"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/mytarget"
	"ru/kovardin/getapp/app/modules/mediation/services/parser/yandex"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/app/utils/admin/components"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

func init() {
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure), fx.Invoke(func(m *Module) {}))
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		// repos
		database.NewRepository[models.Placement],
		database.NewRepository[models.Network],
		database.NewRepository[models.Unit],
		database.NewRepository[models.Cpm],
		database.NewRepository[models.Impression],
		// custom repos
		repos.New,
		// handlers
		handlers.NewNetworks,
		handlers.NewPlacements,
		handlers.NewAuction,
		handlers.NewImpressions,
		// parser
		mytarget.New,
		yandex.New,
		cpa.New,
		bigo.New,
		parser.New,
		// rotation
		bidding.New,
	))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	networks := pb.Model(&models.Network{})
	networks.Listing("ID", "ApplicationId", "Name", "Active", "CreatedAt").
		Field("ApplicationId").
		Label("Application").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := applications.Application{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})

	networks.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:      "application",
				Label:    "Application",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `application_id %s ?`,
			},
		}
	})

	networks.Editing().
		Field("ApplicationId").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Network)
		return web.Portal(components.Dropdown[applications.Application](
			db.DB(),
			c.ApplicationId,
			"Application",
			"Name",
			"ApplicationId",
		)).Name("applications")
	})

	placements := pb.Model(&models.Placement{})
	placements.Listing("ID", "ApplicationId", "Name", "Format", "Active", "CreatedAt").
		Field("ApplicationId").
		Label("Application").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := applications.Application{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	placements.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:      "application",
				Label:    "Application",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `application_id %s ?`,
			},
		}
	})

	placements.Editing().
		Field("ApplicationId").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Placement)
		return web.Portal(components.Dropdown[applications.Application](
			db.DB(),
			c.ApplicationId,
			"Application",
			"Name",
			"ApplicationId",
		)).Name("applications")
	})

	placements.Editing().
		Field("Format").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Placement)
		return web.Portal(components.DropdownList(
			[]components.DropdownListItem{
				{
					Name:  "Interstitial",
					Value: models.UnitFormatInterstitial,
				},
				{
					Name:  "Banner",
					Value: models.UnitFormatBanner,
				},
				{
					Name:  "Native",
					Value: models.UnitFormatNative,
				},
			},
			c.Format,
			"Format",
			"Format",
		)).Name("formats")
	})

	units := pb.Model(&models.Unit{})
	units.Listing("ID", "PlacementId", "NetworkId", "Name", "Unit", "Data", "Active", "CreatedAt").
		Field("PlacementId").
		Label("Placement").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Placement{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	units.Listing().
		Field("NetworkId").
		Label("Network").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Network{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})

	units.Editing().
		Field("NetworkId").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Unit)
		return web.Portal(components.Dropdown[models.Network](
			db.DB(),
			c.NetworkId,
			"Network",
			"Name",
			"NetworkId",
		)).Name("networks")
	})

	units.Editing().
		Field("PlacementId").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Unit)
		return web.Portal(components.Dropdown[models.Placement](
			db.DB(),
			c.PlacementId,
			"Placement",
			"Name",
			"PlacementId",
		)).Name("placements")
	})

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "mediation",
		Subcommands: cli.Commands{
			&cli.Command{
				Name: "mt",
				Action: func(c *cli.Context) error {

					return nil
				},
			},

			&cli.Command{
				Name: "cpms",
				Action: func(c *cli.Context) error {
					setup(c, fx.Invoke(func(log *logger.Logger, cpms *repos.Cpms) {

						to := time.Now()
						from := to.Add(-time.Hour * 24 * 3)

						cpmsByNetworks, err := cpms.CpmsByNetwork(from, to)

						if err != nil {
							log.Info("error on load cpms by network", zap.Error(err))
						}

						for _, item := range cpmsByNetworks {
							fmt.Printf("network: %d, amount: %f\n", item.Network, item.Cpm)
						}

						_ = cpmsByNetworks
					}), fx.NopLogger).Run()

					return nil

				},
			},
		},
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	config      config.Config
	parser      *parser.Parser
	auction     *handlers.Auction
	placements  *handlers.Placements
	impressions *handlers.Impressions
	networks    *handlers.Networks
}

func New(
	lc fx.Lifecycle,
	config config.Config,
	parser *parser.Parser,
	auction *handlers.Auction,
	placements *handlers.Placements,
	impressions *handlers.Impressions,
	networks *handlers.Networks,
) *Module {
	m := &Module{
		config:      config,
		parser:      parser,
		auction:     auction,
		placements:  placements,
		impressions: impressions,
		networks:    networks,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			m.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			m.Stop()
			return nil
		},
	})

	return m
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/mediation", func(r chi.Router) {
		r.Route("/networks/{application}", func(r chi.Router) {
			r.Get("/", m.networks.Networks)
		})

		r.Route("/auction/{placement}", func(r chi.Router) {
			//r.Post("/mediate", m.auction.Mediate)
			r.Post("/bid", m.auction.Bid)
		})

		r.Route("/placements/{placement}", func(r chi.Router) {
			r.Get("/", m.placements.Placement)
		})

		r.Route("/impressions/{placement}", func(r chi.Router) {
			r.Post("/impression", m.impressions.Impression)
		})
	})
}

func (m *Module) Start() {
	if !m.config.Active {
		return
	}
	m.parser.Start()
}

func (m *Module) Stop() {
	if !m.config.Active {
		return
	}
	m.parser.Stop()
}
