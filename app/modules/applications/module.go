package applications

import (
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
	"ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Providers = append(modules.Providers, fx.Provide(
		New,
		database.NewRepository[models.Application],
	))
	modules.Invokes = append(modules.Invokes, fx.Invoke(Configure))
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {
	pb.Model(&models.Application{}).Listing("ID", "Name", "Bundle", "ApiToken", "VkAuthToken", "CreatedAt")

	server.Routers(module)
}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "applications",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("applications")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

type Module struct {
}

func New() *Module {
	return &Module{}
}

func (m *Module) Routes(r chi.Router) {

}
