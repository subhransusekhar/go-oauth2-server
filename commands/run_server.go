package commands

import (
	"net/http"
	"time"

	"github.com/RichardKnop/go-oauth2-server/health"
	"github.com/RichardKnop/go-oauth2-server/oauth"
	"github.com/RichardKnop/go-oauth2-server/web"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"
)

// RunServer runs the app
func RunServer() error {
	cnf, db, err := initConfigDB(true, true)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialise the health service
	healthService := health.NewService(db)

	// Initialise the oauth service
	oauthService := oauth.NewService(cnf, db)

	// Initialise the web service
	webService := web.NewService(cnf, oauthService)

	// Start a classic negroni app
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(gzip.Gzip(gzip.DefaultCompression))
	app.Use(negroni.NewStatic(http.Dir("public")))

	// Create a router instance
	router := mux.NewRouter()

	// Add routes for the health service (healthcheck endpoint)
	health.RegisterRoutes(router, healthService)

	// Add routes for the oauth service (REST tokens endpoint)
	oauth.RegisterRoutes(router, oauthService)

	// Add routes for the web package (register, login authorize web pages)
	web.RegisterRoutes(router, webService)

	// Set the router
	app.UseHandler(router)

	// Run the server on port 8080, gracefully stop on SIGTERM signal
	graceful.Run(":8080", 5*time.Second, app)

	return nil
}
