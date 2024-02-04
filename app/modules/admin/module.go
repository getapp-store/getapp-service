package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	plogin "github.com/qor5/admin/login"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/login"
	"github.com/theplant/htmlgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
	"ru/kovardin/getapp/app/modules/admin/config"
	"ru/kovardin/getapp/app/modules/admin/models"
	h "ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		Login,
		database.NewRepository[models.Setting],
	))
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure))
}

func Login(pb *presets.Builder, database *database.Database) *login.Builder {
	db := database.DB()

	lb := plogin.New(pb).
		DB(db).
		UserModel(&models.Admin{}).
		Secret("123"). // @todo: change
		TOTP(false).
		HomeURLFunc(func(r *http.Request, user interface{}) string {
			return "/admin"
		})

	return lb
}

func Configure(pb *presets.Builder, database *database.Database, module *Module, server *h.Server) {
	db := database.DB()

	pb.
		URIPrefix("/admin").
		BrandTitle("GetApp").
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = vuetify.VContainer(
				htmlgo.H1("Home"),
				htmlgo.P().Text("Change your home page here"))
			return
		})

	admin := pb.Model(&models.Admin{}).MenuIcon("admin_panel_settings")
	admin.Listing("ID", "Account", "CreatedAt")
	admin.Editing("Account")

	setting := pb.Model(&models.Setting{}).MenuIcon("settings")
	setting.Listing("ID", "Key", "Value")

	pb.MenuOrder(
		"Applications",
		pb.MenuGroup("Users").SubItems(
			"Users",
			"Auth",
		),

		pb.MenuGroup("Billing").SubItems(
			"Products",
			"Payments",
		).Icon("attach_money"),

		pb.MenuGroup("Tracking").SubItems(
			"dashboard",
			"Trackers",
			"Conversions",
		).Icon("flag"),
		pb.MenuGroup("Boosty").SubItems(
			"Blogs",
			"Subscriptions",
			"Subscribers",
		).Icon("rocket"),
		pb.MenuGroup("Warehouse").SubItems(
			"Items",
		).Icon("warehouse"),
		pb.MenuGroup("Lokalize").SubItems(
			"Languages",
			"Phrases",
		).Icon("translate"),
		pb.MenuGroup("Landings").SubItems(
			"Landings",
			"Pages",
		).Icon("public"),
		pb.MenuGroup("Mediation").SubItems(
			"Networks",
			"Placements",
			"Units",
		).Icon("ads_click"),
		"Admins",
		"Settings",
	)

	// admin settings
	a := &models.Admin{}
	db.First(&a)
	a.Account = module.config.Username
	a.Password = module.config.Password
	a.EncryptPassword()

	// serve admin
	if err := db.Save(a).Error; err != nil {
		log.Print(err)
	}

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "admin",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {
				fmt.Println("admin")
				os.Exit(0)
			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
	db     *database.Database
	pb     *presets.Builder
	lb     *login.Builder
	config config.Config
}

func New(pb *presets.Builder, lb *login.Builder, config config.Config, db *database.Database) *Module {
	return &Module{
		db:     db,
		pb:     pb,
		lb:     lb,
		config: config,
	}
}

func (m *Module) Routes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(m.lb.Middleware())

		r.Mount("/", m.pb)
	})

	mux := http.NewServeMux()
	m.lb.Mount(mux)

	r.Mount("/", mux)
}
