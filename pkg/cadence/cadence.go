package cadence

import (
	"context"
	"fmt"
	"time"

	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	service "go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/fx"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"

	"ru/kovardin/getapp/pkg/logger"
)

type Config struct {
	Host   string
	Domain string
	//	Tasks - is the task list name you use to identify your client worker, also
	//	           identifies group of workflow and activity implementations that are
	//	           hosted by a single worker process
	Tasks string // "getapp-worker-tasklist"

	Client  string // "getapp-worker"
	Service string // "cadence-frontend"
}

type Cadence struct {
	log     *zap.Logger
	config  Config
	worker  worker.Worker
	service service.Interface
	client  client.Client
}

func New(config Config, lc fx.Lifecycle, log *logger.Logger) *Cadence {
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: config.Client,
		Outbounds: yarpc.Outbounds{
			config.Service: {
				Unary: grpc.NewTransport().NewSingleOutbound(config.Host),
			},
		},
	})

	if err := dispatcher.Start(); err != nil {
		panic(fmt.Errorf("failed to start dispatcher, %w", err))
	}

	cfg := dispatcher.ClientConfig(config.Service)

	service := compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(cfg),
		apiv1.NewWorkflowAPIYARPCClient(cfg),
		apiv1.NewWorkerAPIYARPCClient(cfg),
		apiv1.NewVisibilityAPIYARPCClient(cfg),
	)

	options := worker.Options{
		Logger:       log,
		MetricsScope: tally.NewTestScope(config.Tasks, map[string]string{}),
	}

	worker := worker.New(
		service,
		config.Domain,
		config.Tasks,
		options,
	)

	client := client.NewClient(
		service,
		config.Domain,
		&client.Options{},
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			return worker.Start()
		},
		OnStop: func(ctx context.Context) error {
			worker.Stop()
			return nil
		},
	})

	return &Cadence{
		log:     log,
		config:  config,
		worker:  worker,
		service: service,
		client:  client,
	}
}

func (c *Cadence) RegisterWorkflow(w any, name string) {
	c.worker.RegisterWorkflowWithOptions(w, workflow.RegisterOptions{
		Name: name,
	})
}
func (c *Cadence) RegisterActivity(a any, name string) {
	c.worker.RegisterActivityWithOptions(a, activity.RegisterOptions{
		Name: name,
	})
}

func (c *Cadence) StartWorkflow(id string, workflow any, input string, cron string) {
	ctx := context.Background()

	we, err := c.client.StartWorkflow(ctx, client.StartWorkflowOptions{
		ID:                              id,
		TaskList:                        c.config.Tasks,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
		CronSchedule:                    cron, //"* * * * *",
	}, workflow, input)

	if err != nil {
		c.log.Error("error to start workflow", zap.Error(err))
		return
	}

	c.log.Info("started workflow", zap.Any("workflow", we))
}
