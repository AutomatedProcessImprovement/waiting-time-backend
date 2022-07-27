package app

import (
	"encoding/json"
	"errors"
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

// swagger:operation GET /jobs/{id} getJob
//
// Get a single job.
//
// ---
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Parameters:
//   - name: id
//     in: path
//     description: Job's ID
//     required: true
//     type: string
//
// Responses:
//   default:
//     schema:
//       $ref: '#/definitions/ApiResponseError'
//   200:
//     schema:
//       $ref: '#/definitions/ApiSingleJobResponse'
func GetJobByID(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiResponse model.ApiSingleJobResponse

		vars := mux.Vars(r)
		id := vars["id"]

		job := app.queue.FindByID(id)
		apiResponse.Job = job

		if job == nil {
			var apiResponse model.ApiResponseError
			apiResponse.Error = fmt.Sprintf("job with id %s not found", id)
			reply(w, http.StatusNotFound, apiResponse, app.logger)
			return
		}

		reply(w, http.StatusOK, apiResponse, app.logger)
	}
}

// swagger:operation GET /jobs/{id}/cancel cancelJob
//
// Cancel processing of a job.
//
// ---
// Produces:
//   - application/json
//
// Parameters:
//   - name: id
//     in: path
//     description: Job's ID
//     required: true
//     type: string
//
// Responses:
//   default:
//     schema:
//       $ref: '#/definitions/ApiResponseError'
//   200:
//     schema:
//       $ref: '#/definitions/ApiSingleJobResponse'
func CancelJobByID(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		job := app.queue.FindByID(id)
		if job == nil {
			var apiResponse model.ApiResponseError
			apiResponse.Error = fmt.Sprintf("job with id %s not found", id)
			reply(w, http.StatusNotFound, apiResponse, app.logger)
			return
		}

		if job.Status == model.JobStatusPending {
			job.SetStatus(model.JobStatusFailed)
			job.SetError(errors.New("job cancelled by user"))
			reply(w, http.StatusOK, model.ApiSingleJobResponse{Job: job}, app.logger)
			return
		}

		if job.Status == model.JobStatusRunning {
			// NOTE: this works because we have only one running job at a time and this cancel func cancels whichever
			// job is running at the moment independent of the job ID
			app.analysisCancelFunc()

			job.SetStatus(model.JobStatusFailed)
			reply(w, http.StatusOK, model.ApiSingleJobResponse{Job: job}, app.logger)
			return
		}

		reply(w, http.StatusBadRequest, model.ApiResponseError{Error: "job cannot be cancelled"}, app.logger)
	}
}

// swagger:operation POST /jobs postJob
//
// Submit a job for analysis. The endpoint accepts JSON and CSV request bodies.
//
// ---
// Consumes:
//   - application/json
//   - text/csv
//
// Produces:
//   - application/json
//
// Parameters:
//   - name: Body
//     in: body
//     description: Description of a job
//     required: true
//     schema:
//       $ref: '#/definitions/ApiRequest'
//
// Responses:
//   default:
//     schema:
//       $ref: '#/definitions/ApiResponseError'
//   200:
//     schema:
//       $ref: '#/definitions/ApiSingleJobResponse'
func PostJob(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the event log from the request body
		if r.Header.Get("Content-Type") != "application/json" {
			PostJobFromBody(app)(w, r)
			return
		}

		// Create a new job from the JSON request

		var apiRequest model.ApiRequest

		if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
			message := fmt.Sprintf("invalid request body; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		job, err := model.NewJob(apiRequest.EventLogURL_, apiRequest.CallbackEndpointURL_, app.config.ResultsDir)
		if err != nil {
			message := fmt.Sprintf("cannot create a job; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = job.Validate(); err != nil {
			message := fmt.Sprintf("invalid job; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = app.AddJob(job); err != nil {
			message := fmt.Sprintf("failed to add a job to the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		apiResponse := model.ApiSingleJobResponse{Job: job}
		reply(w, http.StatusCreated, apiResponse, app.logger)
	}
}

func PostJobFromBody(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, err := app.newJobFromRequestBody(r.Body)
		if err != nil {
			message := fmt.Sprintf("failed to create a job from the request body; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = job.Validate(); err != nil {
			message := fmt.Sprintf("invalid job; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = app.AddJob(job); err != nil {
			message := fmt.Sprintf("failed to add a job to the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		apiResponse := model.ApiSingleJobResponse{Job: job}
		reply(w, http.StatusCreated, apiResponse, app.logger)
	}
}

// swagger:route GET /jobs listJobs
//
// List all jobs.
//
// ---
// Responses:
//   default: ApiJobsResponse
func GetJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiResponse := model.ApiJobsResponse{Jobs: app.queue.Jobs}
		reply(w, http.StatusOK, apiResponse, app.logger)
	}
}

func reply(w http.ResponseWriter, statusCode int, response interface{}, logger *log.Logger) {
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
