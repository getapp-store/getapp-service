package deployer

import (
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/presets"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
	"ru/kovardin/getapp/app/servers/http"
	"ru/kovardin/getapp/pkg/database"
)

func init() {
	modules.Commands = append(modules.Commands, Command)
	modules.Modules = append(modules.Modules, Deployer)
}

var Deployer = fx.Module("deployer",
	fx.Provide(
		New,
	),
	fx.Invoke(Configure),
)

type Module struct {
}

func Configure(pb *presets.Builder, db *database.Database, module *Module, server *http.Server) {

}

func Command(setup func(*cli.Context, ...fx.Option) *fx.App) *cli.Command {
	return &cli.Command{
		Name: "deploy",
		Action: func(c *cli.Context) error {
			setup(c, fx.Invoke(func() {

				fmt.Println("deploy")
				os.Exit(0)

			}), fx.NopLogger).Run()

			return nil
		},
	}
}

func New() *Module {
	return &Module{}
}

func (m *Module) Routes(r chi.Router) {

}

func (m *Module) Setup(presets *presets.Builder) {

}
