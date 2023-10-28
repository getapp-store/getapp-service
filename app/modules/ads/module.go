package ads

import (
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
)

func init() {
	modules.Providers = append(modules.Providers, fx.Provide(New))
}

type Module struct {
}

func New() *Module {
	return &Module{}
}
