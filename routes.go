package main

import (
	"net/http"
)

// AddRoutes for api creates routes
func (app *App) AddRoutes() {

	//Project API
	app.AddRoute("POST", "/project", app.AddProjectItem)
	app.AddRoute("GET", "/project", app.GetProjectItems)
	app.AddRoute("GET", "/project/{id}", app.FindProjectItem)
	app.AddRoute("DELETE", "/project/{id}", app.DeleteProjectItem)

	//News API
	app.AddRoute("POST", "/news", app.AddNewsItem)
	app.AddRoute("GET", "/news", app.GetNewsItems)
	app.AddRoute("GET", "/news/{id}", app.FindNewsItem)
	app.AddRoute("DELETE", "/news/{id}", app.DeleteNewsItem)

	//Job API
	app.AddRoute("POST", "/job", app.AddJobApplications)
	app.AddRoute("GET", "/job", app.GetJobApplications)
	app.AddRoute("GET", "/job/{id}", app.FindJobApplicationByID)
	app.AddRoute("DELETE", "/job/{id}", app.DeleteJob)

	//Relation API
	app.AddRoute("POST", "/relation", app.AddRelation)

	//Health Check Status
	app.AddRoute("GET", "/health", app.HealthCheck)
}

//HealthCheck checks application status
func (app *App) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	health := map[string]string{}
	err := app.db.Ping()
	if err != nil {
		health["database"] = "down"
		health["error"] = err.Error()
		health["status"] = "down"
	} else {
		health["database"] = "up"
		health["status"] = "up"
	}
	app.RenderJson(writer, http.StatusOK, health)
}
