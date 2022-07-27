package app

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	PathPrefix  string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (app *Application) initializeRouter() {
	var routes = Routes{
		Route{
			"GetSwaggerJSON",
			"GET",
			"/swagger.json",
			"",
			GetSwaggerJSON(app),
		},

		Route{
			"Assets",
			"GET",
			"",
			"/assets/",
			StaticAssets(app),
		},

		Route{
			"CancelJobByID",
			"GET",
			"/jobs/{id}/cancel",
			"",
			CancelJobByID(app),
		},

		Route{
			"GetJobByID",
			"GET",
			"/jobs/{id}",
			"",
			GetJobByID(app),
		},

		Route{
			"PostJob",
			"POST",
			"/jobs",
			"",
			PostJob(app),
		},

		Route{
			"GetJobs",
			"GET",
			"/jobs",
			"",
			GetJobs(app),
		},

		Route{
			"Index",
			"GET",
			"/",
			"",
			Index,
		},
	}

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = EnableCORS(Logger(app, handler, route.Name))

		if route.PathPrefix != "" {
			router.
				Methods(route.Method, "OPTIONS").
				PathPrefix(route.PathPrefix).
				Name(route.Name).
				Handler(handler)
		} else {
			router.
				Methods(route.Method, "OPTIONS").
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}

	app.router = router
}
