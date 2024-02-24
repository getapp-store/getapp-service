package boosty

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"kovardin.ru/projects/boosty"
	"kovardin.ru/projects/boosty/auth"
	"kovardin.ru/projects/boosty/request"

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/boosty/config"
	"ru/kovardin/getapp/app/modules/boosty/handlers"
	"ru/kovardin/getapp/app/modules/boosty/models"
	"ru/kovardin/getapp/app/modules/boosty/workflow"
	"ru/kovardin/getapp/app/modules/boosty/workflow/parser"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/app/utils/admin/components"
	"ru/kovardin/getapp/pkg/cadence"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		database.NewRepository[models.Blog],
		database.NewRepository[models.Subscription],
		database.NewRepository[models.Subscriber],
		handlers.NewSubscribers,
		handlers.NewSubscriptions,
		handlers.NewBlogs,

		// cadence
		workflow.New,
		parser.New,
	))
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure), fx.Invoke(func(m *Module) {}))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	blogs := pb.Model(&models.Blog{})
	blogs.Listing("ID", "ApplicationID", "Title", "Name", "Url", "Token", "Active", "CreatedAt").
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

	blogs.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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

	blogs.Editing().
		Field("ApplicationID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Blog)
		return web.Portal(components.Dropdown[applications.Application](
			db.DB(),
			c.ApplicationID,
			"Application",
			"Name",
			"ApplicationID",
		)).Name("applications")
	})

	subscriptions := pb.Model(&models.Subscription{})
	subscriptions.Listing("ID", "External", "BlogID", "Title", "Amount", "Active", "CreatedAt").
		Field("BlogID").
		Label("Blog").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Blog{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Title))
		})

	subscriptions.Listing().
		Field("Amount").
		Label("Amount").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			raw, _ := field.Value(obj).(int)

			amount := fmt.Sprintf("%.2f", float64(raw)/100)

			return htmlgo.Td(htmlgo.Text(amount))
		})

	subscriptions.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:      "blog",
				Label:    "Blog",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `blog_id %s ?`,
			},
		}
	})

	subscribers := pb.Model(&models.Subscriber{})
	subscribers.Listing("ID", "External", "BlogID", "SubscriptionID", "Name", "Email", "Active", "CreatedAt").
		Field("BlogID").
		Label("Blog").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Blog{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Title))
		})

	subscribers.Listing().
		Field("SubscriptionID").
		Label("Subscription").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Subscription{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Title))
		})

	subscribers.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:      "blog",
				Label:    "Blog",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `blog_id %s ?`,
			},
			{
				Key:      "subscription",
				Label:    "Subscription",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `subscription_id %s ?`,
			},
		}
	})

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "boosty",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "token",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "blog",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func(log *logger.Logger) {
				// create api
				token := auth.Info{}
				if err := json.Unmarshal([]byte(c.String("token")), &token); err != nil {
					log.Error("error on parse boosty token", zap.Error(err))
					return
				}

				a, err := auth.New(
					auth.WithInfo(auth.Info{}),
					auth.WithInfoUpdateCallback(func(info auth.Info) {
						// todo
					}),
				)

				rq, err := request.New(request.WithAuth(a))
				if err != nil {
					log.Error("error on prepare boosty lib request", zap.Error(err))
					return
				}

				b, err := boosty.New(c.String("blog"), boosty.WithRequest(rq))
				if err != nil {
					log.Error("error on prepare boosty lib", zap.Error(err))
					return
				}

				v := url.Values{}
				v.Add("offset", "0")
				v.Add("limit", "2")
				v.Add("order", "gt")
				v.Add("sort_by", "on_time")

				fmt.Println("subscriptions:")
				ss, err := b.Subscriptions(v)
				if err != nil {
					log.Error("error on get subscriptions", zap.Error(err))
					os.Exit(1)
				}

				t := table.NewWriter()
				for _, s := range ss.Data {
					t.AppendRow(table.Row{s.ID, s.Name, s.Price})
				}
				fmt.Println(t.Render())

				fmt.Println()

				v = url.Values{}
				v.Add("offset", "0")
				v.Add("limit", "2")
				v.Add("order", "gt")
				v.Add("sort_by", "on_time")

				fmt.Println("subscribers:")
				uu, err := b.Subscribers(v)
				if err != nil {
					log.Error("error on get subscribers", zap.Error(err))
					os.Exit(1)
				}

				t = table.NewWriter()
				for _, s := range uu.Data {
					t.AppendRow(table.Row{s.ID, s.Name, s.Email})
				}
				fmt.Println(t.Render())

				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	config        config.Config
	cadence       *cadence.Cadence
	workflow      *workflow.Workflow
	subscriptions *handlers.Subscriptions
	subscribers   *handlers.Subscribers
	blogs         *handlers.Blogs
}

func New(
	lc fx.Lifecycle,
	config config.Config,
	cadence *cadence.Cadence,
	workflow *workflow.Workflow,
	subscriptions *handlers.Subscriptions,
	subscribers *handlers.Subscribers,
	blogs *handlers.Blogs,
) *Module {
	m := &Module{
		config:        config,
		cadence:       cadence,
		workflow:      workflow,
		subscriptions: subscriptions,
		subscribers:   subscribers,
		blogs:         blogs,
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
	m.cadence.StartWorkflow(m.config.Workflow, m.workflow.Execute, "boosty", m.config.Cron)
}

func (m *Module) Stop() {
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/boosty", func(r chi.Router) {
		r.Route("/{application}", func(r chi.Router) {
			r.Get("/subscriptions", m.subscriptions.Subscriptions)
			r.Get("/subscriber/{external}", m.subscribers.Subscriber)
			r.Get("/blog", m.blogs.Blog)
		})
	})
}
