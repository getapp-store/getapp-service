package modules

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var (
	Commands  = []func(func(*cli.Context, ...fx.Option) *fx.App) *cli.Command{}
	Providers = []fx.Option{} // global providers for all commands
	Invokes   = []fx.Option{} // invoke only for server command
)
