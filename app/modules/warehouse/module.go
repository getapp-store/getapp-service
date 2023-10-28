package warehouse

import (
	"fmt"
	"log"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/auth"
	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/warehouse/handlers"
	"ru/kovardin/getapp/app/modules/warehouse/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/utils"
)

func init() {
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure))
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		handlers.NewItems,
		database.NewRepository[models.Item],
	))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	items := pb.Model(&models.Item{})
	items.Listing("ID", "ApplicationID", "Key", "Value", "Active", "CreatedAt").
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

	items.Listing().
		Field("Value").
		Label("Value").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			val, _ := field.Value(obj).(string)

			return htmlgo.Td(htmlgo.Text(utils.Substring(val, 0, 20)))
		})

	items.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "warehouse",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("warehouse")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	items *handlers.Items
	auth  *auth.Auth
}

func New(items *handlers.Items, auth *auth.Auth) *Module {
	return &Module{
		items: items,
		auth:  auth,
	}
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/warehouse", func(r chi.Router) {
		r.Route("/{application}", func(r chi.Router) {
			//r.Use(m.auth.AppAuthorize())

			r.Get("/search", m.items.Search)
			r.Get("/item/{key}", m.items.Item)
		})
	})
}
