package app

import (
	"encoding/json"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	_ "embed"
)

func Index(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello World!")
}

func StaticAssets(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logger.Printf("Serving %s", r.URL.Path[1:])
		http.ServeFile(w, r, r.URL.Path[1:])
	}
}

//go:embed spec/swagger.json
var swaggerJSON string

func GetSwaggerJSON(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, swaggerJSON)
	}
}

func GetJobByID(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiResponse model.ApiResponse

		vars := mux.Vars(r)
		id := vars["id"]

		job := app.queue.FindByID(id)
		apiResponse.Job = job

		if job == nil {
			apiResponse.Error_ = &model.ApiResponseError{Message: "job not found"}
			reply(w, http.StatusNotFound, apiResponse, app.logger)
			return
		}

		reply(w, http.StatusOK, apiResponse, app.logger)
	}
}

func PostJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiRequest model.ApiRequest
		var apiResponse model.ApiResponse

		if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
			message := fmt.Sprintf("invalid request body; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponse{Error_: &model.ApiResponseError{Message: message}}, app.logger)
			return
		}

		job, err := model.NewJob(apiRequest.EventLog, apiRequest.CallbackEndpoint, app.config.ResultsDir)
		if err != nil {
			message := fmt.Sprintf("cannot create a job; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponse{Error_: &model.ApiResponseError{Message: message}}, app.logger)
			return
		}

		if err = job.Validate(); err != nil {
			message := fmt.Sprintf("invalid job; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponse{Error_: &model.ApiResponseError{Message: message}}, app.logger)
			return
		}

		if err = app.AddJob(job); err != nil {
			message := fmt.Sprintf("failed to add a job to the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponse{Error_: &model.ApiResponseError{Message: message}}, app.logger)
			return
		}

		apiResponse.Job = job
		reply(w, http.StatusCreated, apiResponse, app.logger)
	}
}

// swagger:route GET /jobs listJobs
//
// List all jobs
//
// Responses:
//   default: ApiResponse
func GetJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiResponse model.ApiResponse

		apiResponse.Jobs = app.queue.Jobs

		reply(w, http.StatusOK, apiResponse, app.logger)
	}
}

func reply(w http.ResponseWriter, statusCode int, response model.ApiResponse, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	checkError(err, "failed to encode ApiResponse", logger)
}

func checkError(err error, message string, logger *log.Logger) {
	if err == nil {
		return
	}
	logger.Printf("%s; %s", message, err)
}
