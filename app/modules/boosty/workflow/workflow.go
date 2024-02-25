package workflow

import (
	"time"

	w "go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/boosty/workflow/parser"
	"ru/kovardin/getapp/pkg/cadence"
)

type Workflow struct {
	parser *parser.Parser
}

func New(cadence *cadence.Cadence, parser *parser.Parser) *Workflow {
	workflow := &Workflow{
		parser: parser,
	}

	cadence.RegisterWorkflow(workflow.Execute, "main.boosty")
	cadence.RegisterActivity(parser.Execute, "boosty.parser")

	return workflow
}

func (wr *Workflow) Execute(ctx w.Context, name string) error {
	options := w.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}

	ctx = w.WithActivityOptions(ctx, options)

	log := w.GetLogger(ctx)
	log.Info("boosty workflow started")
	var result string

	if err := w.ExecuteActivity(ctx, wr.parser.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	log.Info("boosty workflow completed", zap.String("result", result))

	return nil
}
