package temporal

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// WorkerConfig configures a worker. Zero concurrency values fall back to sane
// defaults.
type WorkerConfig struct {
	Client                       client.Client
	TaskQueue                    string
	MaxConcurrentWorkflows       int
	MaxConcurrentActivities      int
	MaxConcurrentLocalActivities int
	Logger                       *zap.Logger
}

// WorkerBuilder registers workflows and activities and builds a worker.
type WorkerBuilder struct {
	cfg        WorkerConfig
	workflows  []interface{}
	activities []interface{}
}

// NewWorkerBuilder starts a builder with defaults applied.
func NewWorkerBuilder(cfg WorkerConfig) *WorkerBuilder {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.MaxConcurrentWorkflows == 0 {
		cfg.MaxConcurrentWorkflows = 100
	}
	if cfg.MaxConcurrentActivities == 0 {
		cfg.MaxConcurrentActivities = 100
	}
	if cfg.MaxConcurrentLocalActivities == 0 {
		cfg.MaxConcurrentLocalActivities = 100
	}
	return &WorkerBuilder{cfg: cfg}
}

// RegisterWorkflows adds workflow functions.
func (b *WorkerBuilder) RegisterWorkflows(wfs ...interface{}) *WorkerBuilder {
	b.workflows = append(b.workflows, wfs...)
	return b
}

// RegisterActivities adds activity functions or structs.
func (b *WorkerBuilder) RegisterActivities(acts ...interface{}) *WorkerBuilder {
	b.activities = append(b.activities, acts...)
	return b
}

// Build validates and constructs the worker.
func (b *WorkerBuilder) Build() (worker.Worker, error) {
	if b.cfg.Client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if b.cfg.TaskQueue == "" {
		return nil, fmt.Errorf("task queue is required")
	}
	w := worker.New(b.cfg.Client, b.cfg.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:      b.cfg.MaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize:  b.cfg.MaxConcurrentWorkflows,
		MaxConcurrentLocalActivityExecutionSize: b.cfg.MaxConcurrentLocalActivities,
	})
	for _, wf := range b.workflows {
		w.RegisterWorkflow(wf)
	}
	for _, act := range b.activities {
		w.RegisterActivity(act)
	}
	b.cfg.Logger.Info("temporal: worker configured",
		zap.String("task_queue", b.cfg.TaskQueue),
		zap.Int("workflows", len(b.workflows)),
		zap.Int("activities", len(b.activities)))
	return w, nil
}

// Run builds the worker and blocks until interrupted.
func (b *WorkerBuilder) Run() error {
	w, err := b.Build()
	if err != nil {
		return err
	}
	b.cfg.Logger.Info("temporal: worker starting", zap.String("task_queue", b.cfg.TaskQueue))
	return w.Run(worker.InterruptCh())
}
