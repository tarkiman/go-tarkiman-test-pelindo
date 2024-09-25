package router

import (
	"github.com/go-chi/chi"
	"github.com/tarkiman/go/internal/handlers"
)

// DomainHandlers is a struct that contains all domain-specific handlers.
type DomainHandlers struct {
	TaskHandler handlers.TaskHandler
}

// Router is the router struct containing handlers.
type Router struct {
	DomainHandlers DomainHandlers
}

// ProvideRouter is the provider function for this router.
func ProvideRouter(domainHandlers DomainHandlers) Router {
	return Router{
		DomainHandlers: domainHandlers,
	}
}

// SetupRoutes sets up all routing for this server.
func (r *Router) SetupRoutes(mux *chi.Mux) {
	mux.Route("/v1", func(rc chi.Router) {
		r.DomainHandlers.TaskHandler.Router(rc)
	})
}
