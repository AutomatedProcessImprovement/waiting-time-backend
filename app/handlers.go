package app

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func GetJobByID(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiResponse ApiResponse

		vars := mux.Vars(r)
		id := vars["id"]

		job := app.queue.FindByID(id)
		apiResponse.Job = job

		if job == nil {
			apiResponse.Error_ = &ApiResponseError{Message: "job not found"}
			reply(w, http.StatusNotFound, apiResponse)
			return
		}

		reply(w, http.StatusOK, apiResponse)
	}
}

func PostJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiRequest ApiRequest
		var apiResponse ApiResponse

		if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
			message := fmt.Sprintf("invalid request body; %s", err)
			reply(w, http.StatusBadRequest, ApiResponse{Error_: &ApiResponseError{Message: message}})
			return
		}

		job, err := NewJob(apiRequest.EventLog, apiRequest.CallbackEndpoint)
		if err != nil {
			message := fmt.Sprintf("cannot create a job; %s", err)
			reply(w, http.StatusBadRequest, ApiResponse{Error_: &ApiResponseError{Message: message}})
			return
		}

		if err = job.Validate(); err != nil {
			message := fmt.Sprintf("invalid job; %s", err)
			reply(w, http.StatusBadRequest, ApiResponse{Error_: &ApiResponseError{Message: message}})
			return
		}

		if err = app.queue.Add(job); err != nil {
			message := fmt.Sprintf("failed to add a job to the queue; %s", err)
			reply(w, http.StatusInternalServerError, ApiResponse{Error_: &ApiResponseError{Message: message}})
			return
		}

		go app.processJob(job)

		apiResponse.Job = job
		reply(w, http.StatusCreated, apiResponse)
	}
}

func GetJobs(app *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiResponse ApiResponse

		apiResponse.Jobs = app.queue.Jobs

		reply(w, http.StatusOK, apiResponse)
	}
}

func reply(w http.ResponseWriter, statusCode int, response ApiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	checkError(err, "failed to encode ApiResponse")
}

func checkError(err error, message string) {
	if err == nil {
		return
	}
	log.Printf("%s; %s", message, err)
}
