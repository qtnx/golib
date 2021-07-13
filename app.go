package golib

import (
	"gitlab.id.vin/vincart/golib/config"
	"gitlab.id.vin/vincart/golib/log"
	"gitlab.id.vin/vincart/golib/pubsub"
	"gitlab.id.vin/vincart/golib/web/middleware"
	"net/http"
)

type Module func(app *App)

type App struct {
	Loader     config.Loader
	Logger     log.Logger
	Publisher  pubsub.Publisher
	Properties *Properties
}

func New(modules ...Module) *App {
	app := App{}
	for _, module := range modules {
		module(&app)
	}
	return &app
}

func (a App) Middleware() []func(next http.Handler) http.Handler {
	return []func(next http.Handler) http.Handler{
		middleware.AdvancedResponseWriter(),
		middleware.RequestContext(),
		middleware.CorrelationId(),
	}
}
