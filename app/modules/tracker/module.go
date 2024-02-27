package tracker

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"log"
	"os"
	"ru/kovardin/getapp/app/modules/tracker/dashboards"

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/tracker/config"
	"ru/kovardin/getapp/app/modules/tracker/handlers"
	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/app/modules/tracker/workflow"
	"ru/kovardin/getapp/app/modules/tracker/workflow/vkads"
	"ru/kovardin/getapp/app/modules/tracker/workflow/yandex"
	server "ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/cadence"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		handlers.NewConversions,
		database.NewRepository[models.Conversion],
		database.NewRepository[models.Tracker],

		// dashboards
		dashboards.NewTracker,

		// cadence
		workflow.New,
		yandex.New,
		vkads.New,
	))
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure), fx.Invoke(func(m *Module) {}))
}

type Dashboard struct{}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *server.Server) {
	tt := pb.Model(&models.Tracker{})
	tt.Listing("ID", "Name", "ApplicationID", "YandexMetricaTracker", "VkTracker", "YandexToken", "Active", "CreatedAt").
		Field("ApplicationID").
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

	tt.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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

	// dashboard
	aa := pb.Model(&Dashboard{}).Label("Dashboard")

	aa.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = htmlgo.Div(
			htmlgo.Iframe().
				Src("/admin/tracker/dashboard").
				Attr("width", "100%", "height", "100%", "frameborder", "no").
				Style("transform-origin: left top; transform: scale(1, 1); pointer-events: none;"),
		).Style("height: 100vh; width: 100%")

		r.PageTitle = "Dashboard"

		return
	})

	cc := pb.Model(&models.Conversion{})
	cc.Listing("ID", "RbClickid", "Yclid", "Fire", "TrackerID", "CreatedAt").
		Field("TrackerID").
		Label("Tracker").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Tracker{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	cc.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		options := []*vuetifyx.SelectItem{
			{Text: "True", Value: "true"},
			{Text: "False", Value: "false"},
		}

		return []*vuetifyx.FilterItem{
			{
				Key:      "fire",
				Label:    "Fire",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `fire %s ?`,
				Options:      options,
			},
			{
				Key:      "tracker",
				Label:    "Tracker",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `tracker_id %s ?`,
			},
		}
	})

	server.Routers(module)
	server.Dashboaders(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "tracker",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("tracker")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	config      config.Config
	conversions *handlers.Conversions
	cadence     *cadence.Cadence
	workflow    *workflow.Workflow
	dashboard   *dashboards.Tracker
}

func New(
	lc fx.Lifecycle,
	config config.Config,
	conversions *handlers.Conversions,
	cadence *cadence.Cadence,
	workflow *workflow.Workflow,
	dashboard *dashboards.Tracker,
) *Module {
	m := &Module{
		config:      config,
		conversions: conversions,
		cadence:     cadence,
		workflow:    workflow,
		dashboard:   dashboard,
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

func (m *Module) Start() {
	m.cadence.StartWorkflow(m.config.Workflow, m.workflow.Execute, "tracker", m.config.Cron)
}

func (m *Module) Stop() {

}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/fire", func(r chi.Router) {
		r.Get("/", m.conversions.Fire)
	})
}

func (m *Module) Dashboards(r chi.Router) {
	r.Get("/tracker/dashboard", m.dashboard.Dashboard)
}
