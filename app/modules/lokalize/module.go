package lokalize

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

	"ru/kovardin/getapp/app/modules"
	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/lokalize/handlers"
	"ru/kovardin/getapp/app/modules/lokalize/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/utils"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Modules = append(modules.Modules, Lokalize)
}

var Lokalize = fx.Module("lokalize",
	fx.Provide(
		New,
		database.NewRepository[models.Language],
		database.NewRepository[models.Phrase],
		handlers.NewLanguages,
		handlers.NewPhrases,
	),
	fx.Invoke(Configure),
)

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	languages := pb.Model(&models.Language{})
	languages.Listing("ID", "ApplicationID", "Name", "Locale", "Active", "CreatedAt").
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

	languages.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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

	phrases := pb.Model(&models.Phrase{})
	phrases.Listing("ID", "LanguageID", "Key", "Value", "Active", "CreatedAt").
		Field("LanguageID").
		Label("Language").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			c := models.Language{}
			cid, _ := field.Value(obj).(uint)
			if err := db.DB().Where("id = ?", cid).Find(&c).Error; err != nil {
				// ignore err in the example
				log.Print(err)
			}
			return htmlgo.Td(htmlgo.Text(c.Name))
		})

	phrases.Listing().
		Field("Value").
		Label("Value").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			val, _ := field.Value(obj).(string)

			return htmlgo.Td(htmlgo.Text(utils.Substring(val, 0, 200)))
		})

	phrases.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:      "language",
				Label:    "Language",
				ItemType: vuetifyx.ItemTypeNumber,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `language_id %s ?`,
			},
		}
	})

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "lokalize",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("lokalize")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	languages *handlers.Languages
	phrases   *handlers.Phrases
}

func New(languages *handlers.Languages, phrases *handlers.Phrases) *Module {
	return &Module{
		languages: languages,
		phrases:   phrases,
	}
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/lokalize", func(r chi.Router) {
		r.Route("/{application}", func(r chi.Router) {
			//r.Use(m.auth.AppAuthorize())

			r.Get("/languages", m.languages.List)
			r.Get("/languages/{locale}.json", m.phrases.List)
		})
	})
}
