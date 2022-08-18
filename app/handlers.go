package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path"
	"strings"

	_ "embed"
)

func Index(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello World!")
}

func StaticAssets(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.TrimPrefix(r.URL.Path, "/assets/")
		assetPath := path.Join(app.config.AssetsDir, filePath)
		http.ServeFile(w, r, assetPath)
	}
}

//go:embed spec/swagger.json
var swaggerJSON string

func SwaggerJSON(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, swaggerJSON)
	}
}

// swagger:operation POST /callback postCallback
//
// Sample endpoint that receives a callback from the analysis service and responds with a 200 OK.
//
// ---
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Parameters:
//   - name: Body
//     in: body
//     description: Callback request
//     required: true
//     schema:
//     $ref: '#/definitions/ApiCallbackRequest'
//
// Responses:
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiResponseError'
//	200:
//	  schema:
//	    $ref: '#/definitions/ApiCallbackRequest'
func SampleCallback(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload model.ApiCallbackRequest

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			message := fmt.Sprintf("invalid request body; %s", err)
			reply(w, http.StatusBadRequest, model.ApiResponseError{Error: message}, app.logger)
			return
		}
		_ = r.Body.Close()

		app.logger.Printf("Received callback %v", payload)
		reply(w, http.StatusOK, payload, app.logger)
	}
}

// swagger:route GET /jobs listJobs
//
// List all jobs.
//
// ---
// Responses:
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiJobsResponse'
func GetJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiResponse := model.ApiJobsResponse{Jobs: app.queue.Jobs}
		reply(w, http.StatusOK, apiResponse, app.logger)
	}
}

// swagger:operation POST /jobs postJob
//
// Submit a job for analysis. The endpoint accepts JSON and CSV request bodies. If the callback URL is provided, a GET
// request with empty body is sent to this endpoint when analysis is complete.
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
//     $ref: '#/definitions/ApiRequest'
//
// Responses:
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiResponseError'
//	200:
//	  schema:
//	    $ref: '#/definitions/ApiSingleJobResponse'
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
		_ = r.Body.Close()

		job, err := model.NewJob(apiRequest.EventLogURL_, apiRequest.CallbackEndpointURL_, apiRequest.ColumnMapping, app.config.ResultsDir)
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
		columnMapping := columnMappingFromRequest(r)

		job, err := app.newJobFromRequestBody(r.Body, columnMapping)
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

// swagger:route DELETE /jobs deleteJobs
//
// Delete all non-running jobs. If a job is running, it returns an error. Cancel the running jobs manually before deleting them.
//
// ---
// Responses:
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiResponseError'
//	200:
//	  schema:
//	    $ref: '#/definitions/ApiJobsResponse'
func DeleteJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := app.queue.Clear()
		if err != nil {
			message := fmt.Sprintf("failed to clear the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = app.SaveQueue(); err != nil {
			message := fmt.Sprintf("failed to save the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		if err = app.LoadQueue(); err != nil {
			message := fmt.Sprintf("failed to load the queue; %s", err)
			reply(w, http.StatusInternalServerError, model.ApiResponseError{Error: message}, app.logger)
			return
		}

		apiResponse := model.ApiJobsResponse{Jobs: app.queue.Jobs}
		reply(w, http.StatusOK, apiResponse, app.logger)
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
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiResponseError'
//	200:
//	  schema:
//	    $ref: '#/definitions/ApiSingleJobResponse'
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
//
//	default:
//	  schema:
//	    $ref: '#/definitions/ApiResponseError'
//	200:
//	  schema:
//	    $ref: '#/definitions/ApiSingleJobResponse'
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

func reply(w http.ResponseWriter, statusCode int, response interface{}, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	checkError(err, "failed to encode JSON response", logger)
}

func checkError(err error, message string, logger *log.Logger) {
	if err == nil {
		return
	}
	logger.Printf("%s; %s", message, err)
}

func columnMappingFromRequest(r *http.Request) map[string]string {
	vars := mapFromRawQuery(r.URL.RawQuery)

	var (
		caseID    string
		activity  string
		resource  string
		startTime string
		endTime   string
		ok        bool
	)

	columnMapping := make(map[string]string)

	caseID, ok = vars["case"]
	if ok {
		columnMapping["case"] = caseID
	}
	activity, ok = vars["activity"]
	if ok {
		columnMapping["activity"] = activity
	}
	resource, ok = vars["resource"]
	if ok {
		columnMapping["resource"] = resource
	}
	startTime, ok = vars["start_timestamp"]
	if ok {
		columnMapping["start_timestamp"] = startTime
	}
	endTime, ok = vars["end_timestamp"]
	if ok {
		columnMapping["end_timestamp"] = endTime
	}

	return columnMapping
}

func mapFromRawQuery(query string) map[string]string {
	queryMap := make(map[string]string)
	for _, pair := range strings.Split(query, "&") {
		values := strings.Split(pair, "=")
		if len(values) == 2 {
			queryMap[values[0]] = values[1]
		}
	}
	return queryMap
}
