package workflow

import (
	"time"

	"go.uber.org/zap"

	w "go.uber.org/cadence/workflow"
	"ru/kovardin/getapp/app/modules/tracker/config"
	"ru/kovardin/getapp/app/modules/tracker/workflow/vkads"
	"ru/kovardin/getapp/app/modules/tracker/workflow/yandex"
	"ru/kovardin/getapp/pkg/cadence"
)

type Workflow struct {
	config  config.Config
	cadence *cadence.Cadence
	yandex  *yandex.Yandex
	vkads   *vkads.Vkads
}

func New(config config.Config, cadence *cadence.Cadence, yandex *yandex.Yandex, vkads *vkads.Vkads) *Workflow {
	workflow := &Workflow{
		config:  config,
		cadence: cadence,
		yandex:  yandex,
		vkads:   vkads,
	}

	cadence.RegisterWorkflow(workflow.Execute, "main.tracker")
	cadence.RegisterActivity(yandex.Execute, "tracker.yandex")
	cadence.RegisterActivity(vkads.Execute, "tracker.vkads")

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

	var result string

	if !wr.config.Active {
		result = "disabled"
		log.Info("tracker workflow disabled", zap.String("result", result))
		return nil
	}

	log.Info("tracker workflow started")

	if err := w.ExecuteActivity(ctx, wr.yandex.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	if err := w.ExecuteActivity(ctx, wr.vkads.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	log.Info("tracker workflow completed", zap.String("result", result))

	return nil
}
