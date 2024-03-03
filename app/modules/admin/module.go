package admin

import (
	"fmt"
	"github.com/qor5/ui/vuetify"
	"log"
	"net/http"
	"os"
	"ru/kovardin/getapp/app/modules/admin/dashboards"

	"github.com/go-chi/chi/v5"
	plogin "github.com/qor5/admin/login"
	"github.com/qor5/admin/presets"
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
	modules.Modules = append(modules.Modules, Admin)
}

var Admin = fx.Module("admin",
	fx.Provide(
		New,
		Login,
		database.NewRepository[models.Setting],
		// dashboards
		dashboards.NewHome,
	),
	fx.Invoke(Configure),
)

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
		BrandFunc(func(ctx *web.EventContext) htmlgo.HTMLComponent {
			return vuetify.VCardText(
				htmlgo.A(htmlgo.H1("GetApp")).Href("/admin"),
			).Class("pa-0")
		}).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = htmlgo.Div(
				htmlgo.Iframe().
					Src("/admin/home/dashboard").
					Attr("width", "100%", "height", "100%", "frameborder", "no").
					Style("transform-origin: left top; transform: scale(1, 1);"),
			).Style("height: 100vh; width: 100%")
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
	server.Dashboaders(module)
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
	home   *dashboards.Home
}

func New(
	pb *presets.Builder,
	lb *login.Builder,
	config config.Config,
	db *database.Database,
	home *dashboards.Home,
) *Module {
	return &Module{
		db:     db,
		pb:     pb,
		lb:     lb,
		config: config,
		home:   home,
	}
}

func (m *Module) Routes(r chi.Router) {

}

func (m *Module) Dashboards(r chi.Router) {
	r.Get("/home/dashboard", m.home.Dashboard)
}
