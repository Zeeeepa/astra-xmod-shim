package bootstrap

import (
	"astron-xmod-shim/api/server"
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/internal/core/orchestrator"
	"astron-xmod-shim/internal/core/reconciler"
	"astron-xmod-shim/internal/core/shimlet"
	_ "astron-xmod-shim/internal/core/shimlet/shimlets"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/workqueue"
	"astron-xmod-shim/pkg/log"
	"fmt"
)

func Init(configPath string) error {
	// init config
	config.SetConfigPath(configPath)
	cfg := config.Get()

	// init log
	if err := log.Init(&cfg.Log); err != nil {
		return fmt.Errorf("log configured error: %w", err) // Log initialization failed, cannot use log output
	}
	log.Info("log configured", "cfg: ", cfg.Log)

	// shimlet registry already initialed from init()
	shimReg := shimlet.Registry
	pipeReg := goal.Registry

	// Initialize clients during bootstrap phase

	// init reconciler
	workerNum := 5
	workQueue := workqueue.New()
	// init specStore - Replace MemoryStore with EtcdStore
	specStore := spec.NewEtcdStore()
	//specStore := spec.NewMemoryStore()

	rc := reconciler.NewReconciler(specStore, workerNum, workQueue)

	// Initialize global Tracer singleton
	infraShim, _ := shimReg.GetSingleton(cfg.CurrentShimlet)

	// TODO Use shimlet to get service list
	_, _ = infraShim.ListDeployedServices()

	// init orchestrator
	orchestrator.GlobalOrchestrator = orchestrator.NewOrchestrator(shimReg, pipeReg, workQueue, specStore)

	// start reconciler
	rc.Start()

	specStore.ReloadAll(workQueue)

	// 6. Initialize HTTP Server
	if err := server.Init(); err != nil {
		return fmt.Errorf("HTTP Server initialization failed: %w", err)
	}

	return nil
}

func WaitForShutDown() {
	// TODO graceful shutdown logics
}