package tracker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"github.com/wcharczuk/go-chart/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/tracker/handlers"
	"ru/kovardin/getapp/app/modules/tracker/models"
	"ru/kovardin/getapp/app/modules/tracker/services"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		handlers.NewConversions,
		database.NewRepository[models.Conversion],
		database.NewRepository[models.Tracker],
		services.NewUploader,
	))
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure), fx.Invoke(func(m *Module) {}))
}

type Dashboard struct{}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
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
		// https://github.com/go-echarts/go-echarts
		conversions := []models.Conversion{}
		db.DB().Model(models.Conversion{}).Find(&conversions)

		values := map[time.Time]float64{}

		xvalues := []time.Time{}
		yvalues := []float64{}

		for _, conversion := range conversions {
			x := conversion.CreatedAt.Truncate(24 * time.Hour)
			y, _ := values[x]
			values[x] = y + 1
		}

		for x, _ := range values {
			xvalues = append(xvalues, x)
		}

		sort.Slice(xvalues, func(i, j int) bool {
			return xvalues[i].Before(xvalues[j])
		})

		for _, y := range xvalues {
			yvalues = append(yvalues, values[y])
		}

		graph := chart.Chart{
			Series: []chart.Series{
				chart.TimeSeries{
					XValues: xvalues,
					YValues: yvalues,
				},
			},
		}

		buffer := bytes.NewBuffer([]byte{})
		if err = graph.Render(chart.SVG, buffer); err != nil {
			return
		}

		r.Body = vuetify.VContainer(
			vuetify.VRow(
				htmlgo.Div(
					htmlgo.Div(htmlgo.H2("Conversions")).Class("mt-2 col col-12"),
					htmlgo.Div(
						htmlgo.RawHTML(
							strings.Replace(buffer.String(), `width="1024" height="1024"`, `width="100%" height="100%" viewBox="-85 -80 1200 1200"`, -1)),
					).Style("height: 300px;"),
				).Class("col col-6"),
			),
		)

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
	conversions *handlers.Conversions
	uploader    *services.Uploader
}

func New(lc fx.Lifecycle, conversions *handlers.Conversions, uploader *services.Uploader) *Module {
	m := &Module{
		conversions: conversions,
		uploader:    uploader,
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
	m.uploader.Start()
}

func (m *Module) Stop() {
	m.uploader.Stop()
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/fire", func(r chi.Router) {
		r.Get("/", m.conversions.Fire)
	})
}
