// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/nextwavedevs/drop/business/auth"
	"github.com/nextwavedevs/drop/business/data/studio"
	"github.com/nextwavedevs/drop/business/data/user"
	"github.com/nextwavedevs/drop/business/mid"
	"github.com/nextwavedevs/drop/foundation/web"
	"go.mongodb.org/mongo-driver/mongo"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *mongo.Client, options ...func(opts *Options)) http.Handler {

	var opts Options
	for _, option := range options {
		option(&opts)
	}

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	//Register check group
	cg := checkGroup{
		build: build,
		db:    db,
	}
	
	app.HandleDebug(http.MethodGet, "/readiness", cg.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth:  a,
	}

	app.Handle(http.MethodGet, "/v1/users/:page/:rows", ug.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin)) //<== you can't do this if you are not an admin and are not yet authenticated. so he used the get token with his id as kid to generate token
	app.Handle(http.MethodGet, "/v1/users/token/:kid", ug.token)
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/users", ug.create)
	app.Handle(http.MethodPut, "/v1/users/:id", ug.update, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", ug.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	// Register studio endpoints.
	sg := studioGroup{
		studio: studio.New(log, db),
	}

	app.Handle(http.MethodGet, "/v1/studio/:page/:rows", sg.query)
	app.Handle(http.MethodGet, "/v1/studio/:page/:rows/:city", sg.queryByLocation)
	app.Handle(http.MethodGet, "/v1/studio/:id", sg.queryByID)
	app.Handle(http.MethodPost, "/v1/studio", sg.create)
	app.Handle(http.MethodPut, "/v1/studio/:id", sg.update)
	app.Handle(http.MethodDelete, "/v1/studio/:id", sg.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	// Accept CORS 'OPTIONS' preflight requests if config has been provided.
	// Don't forget to apply the CORS middleware to the routes that need it.
	// Example Config: `conf:"default:https://MY_DOMAIN.COM"`
	if opts.corsOrigin != "" {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return nil
		}
		app.Handle(http.MethodOptions, "/*", h)
	}

	return app
}
