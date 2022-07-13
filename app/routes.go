package app

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (app *Application) initializeRouter() {
	var routes = Routes{
		Route{
			"Index",
			"GET",
			"/",
			Index,
		},

		Route{
			"GetJobByID",
			"GET",
			"/jobs/{id}",
			GetJobByID(app),
		},

		Route{
			"PostJobs",
			"POST",
			"/jobs",
			PostJobs(app),
		},

		Route{
			"GetJobs",
			"GET",
			"/jobs",
			GetJobs(app),
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = EnableCORS(Logger(app, handler, route.Name))

		router.
			Methods(route.Method, "OPTIONS").
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	router.
		Name("Asset").
		Methods("GET").
		PathPrefix("/assets/").
		Handler(ServeStatic(app))

	app.router = router
}
