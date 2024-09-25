//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tarkiman/go/configs"
	"github.com/tarkiman/go/infras"
	"github.com/tarkiman/go/internal/domain/task"
	"github.com/tarkiman/go/internal/handlers"
	"github.com/tarkiman/go/transport/http"
	"github.com/tarkiman/go/transport/http/middleware"
	"github.com/tarkiman/go/transport/http/router"
)

// Wiring for configurations.
var configurations = wire.NewSet(
	configs.Get,
)

// Wiring for persistences.
var persistences = wire.NewSet(
	// infras.ProvideMySQLConn,
	infras.ProvideOracleConn,
)

// Wiring for domain Task.
var domainTask = wire.NewSet(
	// TaskService interface and implementation
	task.ProvideTaskServiceImpl,
	wire.Bind(new(task.TaskService), new(*task.TaskServiceImpl)),
	// TaskRepository interface and implementation
	// task.ProvideTaskRepositoryMySQL,
	// wire.Bind(new(task.TaskRepository), new(*task.TaskRepositoryMySQL)),
	task.ProvideTaskRepositoryOracle,
	wire.Bind(new(task.TaskRepository), new(*task.TaskRepositoryOracle)),
)

// Wiring for all domains.
var domains = wire.NewSet(
	domainTask,
)

var authMiddleware = wire.NewSet(
	middleware.ProvideAuthentication,
)

// Wiring for HTTP routing.
var routing = wire.NewSet(
	wire.Struct(new(router.DomainHandlers), "*"),
	handlers.ProvideTaskHandler,
	router.ProvideRouter,
)

// Wiring for everything.
func InitializeService() *http.HTTP {
	wire.Build(
		// configurations
		configurations,
		// persistences
		persistences,
		// middleware
		authMiddleware,
		// domains
		domains,
		// routing
		routing,
		// selected transport layer
		http.ProvideHTTP)
	return &http.HTTP{}
}
