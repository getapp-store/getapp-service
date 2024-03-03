package modules

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var (
	Commands = []func(func(*cli.Context, ...fx.Option) *fx.App) *cli.Command{}
	Modules  = []fx.Option{}
)
