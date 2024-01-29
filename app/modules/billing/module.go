package billing

import (
	"fmt"
	"log"
	"os"
	"ru/kovardin/getapp/app/utils/admin/components"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/auth"
	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/billing/config"
	"ru/kovardin/getapp/app/modules/billing/handlers"
	"ru/kovardin/getapp/app/modules/billing/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure))
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		handlers.NewConfirm,
		handlers.NewPayments,
		handlers.NewProducts,

		database.NewRepository[models.Payment],
		database.NewRepository[models.Product],

		auth.New,
	))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	products := pb.Model(&models.Product{})
	products.Listing("ID", "Name", "Title", "Amount", "ApplicationID", "Active", "CreatedAt").
		Field("ApplicationID").
		Label("Application").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := applications.Application{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	products.Listing().
		Field("Amount").
		Label("Amount").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			raw, _ := field.Value(obj).(int)

			amount := fmt.Sprintf("%.2f", float64(raw)/100)

			return htmlgo.Td(htmlgo.Text(amount))
		})
	products.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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
	products.Editing().
		Field("ApplicationID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
		c := obj.(*models.Product)
		return web.Portal(components.Dropdown[applications.Application](
			db.DB(),
			c.ApplicationID,
			"Application",
			"Name",
			"ApplicationID",
		)).Name("applications")
	})

	payments := pb.Model(&models.Payment{})
	payments.Listing().Field("ProductId").
		Label("Product").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Product{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	payments.Listing().Field("ApplicationId").
		Label("Application").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := applications.Application{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	payments.Listing().
		Field("Amount").
		Label("Amount").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			raw, _ := field.Value(obj).(int)

			amount := fmt.Sprintf("%.2f", float64(raw)/100)

			return htmlgo.Td(htmlgo.Text(amount))
		})
	payments.Listing("ID", "Amount", "Status", "ProductId", "ApplicationId", "UserId", "CreatedAt").
		FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
			return []*vuetifyx.FilterItem{
				{
					Key:      "product",
					Label:    "Product",
					ItemType: vuetifyx.ItemTypeNumber,
					// %s is the condition. e.g. >, >=, =, <, <=, like，
					// ? is the value of selected option
					SQLCondition: `product_id %s ?`,
				},
				{
					Key:      "user",
					Label:    "User",
					ItemType: vuetifyx.ItemTypeNumber,
					// %s is the condition. e.g. >, >=, =, <, <=, like，
					// ? is the value of selected option
					SQLCondition: `user_id %s ?`,
				},
			}
		})
	payments.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		options := []*vuetifyx.SelectItem{
			{Text: "Created", Value: models.PaymentStatusCreated},
			{Text: "Confirm", Value: models.PaymentStatusConfirm},
			{Text: "Success", Value: models.PaymentStatusSuccess},
		}

		return []*vuetifyx.FilterItem{
			{
				Key:      "status",
				Label:    "Status",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `status %s ?`,
				Options:      options,
			},
		}
	})

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "billing",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("billing")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	config   config.Config
	logger   *zap.Logger
	confirm  *handlers.Confirm
	payments *handlers.Payments
	products *handlers.Products
	auth     *auth.Auth
}

func New(
	config config.Config,
	logger *zap.Logger,
	callbacks *handlers.Confirm,
	payments *handlers.Payments,
	products *handlers.Products,
	auth *auth.Auth,
) *Module {
	return &Module{
		config:   config,
		logger:   logger,
		confirm:  callbacks,
		payments: payments,
		products: products,
		auth:     auth,
	}
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/billing", func(r chi.Router) {

		r.Route("/{application}", func(r chi.Router) {
			r.Route("/payments", func(r chi.Router) {
				r.Use(m.auth.UserAuthorize())

				// восстанавливать покупки нужно через авторизацию
				r.Get("/restore", m.payments.Restore)

				// процесс покупки происходить через авторизацию
				r.Get("/purchase", m.payments.Purchase)

				r.Get("/{payment}", m.payments.Payment)
			})

			// список продуктов
			r.Get("/products", m.products.List)

			// экран завершения покупки
			r.Get("/success", m.payments.Success)
		})

		r.Group(func(r chi.Router) {
			r.Use(m.auth.AppAuthorize())

			// подтверждение покупки от yoomoney
			r.Post("/confirm", m.confirm.Hook)
		})

	})
}
