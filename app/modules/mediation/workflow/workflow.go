package workflow

import (
	"ru/kovardin/getapp/app/modules/mediation/config"
	"time"

	w "go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/mediation/workflow/bigo"
	"ru/kovardin/getapp/app/modules/mediation/workflow/mytarget"
	"ru/kovardin/getapp/app/modules/mediation/workflow/yandex"
	"ru/kovardin/getapp/pkg/cadence"
)

type Workflow struct {
	config   config.Config
	yandex   *yandex.Yandex
	mytarget *mytarget.MyTarget
	bigo     *bigo.Bigo
}

func New(
	config config.Config,
	cadence *cadence.Cadence,
	yandex *yandex.Yandex,
	mytarget *mytarget.MyTarget,
	bigo *bigo.Bigo,
) *Workflow {
	workflow := &Workflow{
		config:   config,
		yandex:   yandex,
		mytarget: mytarget,
		bigo:     bigo,
	}

	cadence.RegisterWorkflow(workflow.Execute, "main.ecpms")
	cadence.RegisterActivity(yandex.Execute, "ecpm.yandex")
	cadence.RegisterActivity(mytarget.Execute, "ecpm.mytarget")
	cadence.RegisterActivity(bigo.Execute, "ecpm.bigo")

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
		log.Info("ecpm workflow disabled", zap.String("result", result))
		return nil
	}

	log.Info("ecpm workflow started")

	if err := w.ExecuteActivity(ctx, wr.yandex.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	if err := w.ExecuteActivity(ctx, wr.mytarget.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	if err := w.ExecuteActivity(ctx, wr.bigo.Execute, name).Get(ctx, &result); err != nil {
		log.Error("activity failed", zap.Error(err))
	}

	log.Info("ecpm workflow completed", zap.String("result", result))

	return nil
}
