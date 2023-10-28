package users

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
	"ru/kovardin/getapp/app/modules/users/handlers"
	"ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure))
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		// variants
		handlers.NewVkontakte,
		handlers.NewMail,
		handlers.NewAuthorization,

		database.NewRepository[models.Pincode],
		database.NewRepository[models.User],
		database.NewRepository[models.Auth],
	))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	uu := pb.Model(&models.User{})
	uu.Listing("ID", "ExternalId", "ApiToken", "Email", "ApplicationID", "CreatedAt").
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
	uu.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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

	au := pb.Model(&models.Auth{})
	au.Listing("ID", "Title", "Name", "ApplicationID", "Active", "CreatedAt").
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
	au.Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
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
		Name: "users",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("users")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	vkontakte *handlers.Vkontakte
	mail      *handlers.Mail
	auth      *handlers.Authorization
}

func New(vkontakte *handlers.Vkontakte, mail *handlers.Mail, auth *handlers.Authorization) *Module {
	return &Module{
		vkontakte: vkontakte,
		mail:      mail,
		auth:      auth,
	}
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/v1/users/{application}", func(r chi.Router) {
		r.Get("/choose", m.auth.Choose) // выбор способа авторизации

		r.Route("/vk", func(r chi.Router) {
			r.Get("/login", m.vkontakte.Login)
			r.Get("/auth", m.vkontakte.Auth)
			//r.Get("/success", m.vkontakte.Success) // redirect with token?
		})

		r.Route("/mail", func(r chi.Router) {
			r.Get("/login", m.mail.Login)
			r.Post("/send", m.mail.Send)
			r.Post("/auth", m.mail.Auth)
			r.Get("/success", m.mail.Success)
		})
	})
}
