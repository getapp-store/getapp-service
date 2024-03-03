package landings

import (
	"fmt"
	"log"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/landings/handlers"
	"ru/kovardin/getapp/app/modules/landings/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Modules = append(modules.Modules, Landings)
}

var Landings = fx.Module("landings",
	fx.Provide(
		New,
		handlers.NewPages,
		database.NewRepository[models.Landing],
		database.NewRepository[models.Page],
	),
	fx.Invoke(Configure),
)

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	lgs := pb.Model(&models.Landing{})
	lgs.Listing("ID", "ApplicationID", "Name", "Path", "Active", "CreatedAt").
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

	pages := pb.Model(&models.Page{})
	pages.Listing("ID", "LandingID", "Name", "Path", "Active", "CreatedAt").
		Field("LandingID").
		Label("Landing").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Landing{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})
	pages.Editing().
		Field("Body").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			return vuetify.VTextarea().FieldName(field.Name).Label(field.Label).Value(field.Value(obj))
		})

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "landings",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("landings")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	pages *handlers.Pages
}

func (m *Module) Routes(r chi.Router) {
	r.Get("/app/{landing}/{page}", m.pages.Page)
}

func New(pages *handlers.Pages) *Module {
	return &Module{
		pages: pages,
	}
}
