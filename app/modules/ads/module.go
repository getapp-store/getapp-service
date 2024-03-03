package ads

import (
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/modules"
)

func init() {
	modules.Modules = append(modules.Modules, Ads)
}

var Ads = fx.Module("ads",
	fx.Provide(
		New,
	),
)

type Module struct {
}

func New() *Module {
	return &Module{}
}
