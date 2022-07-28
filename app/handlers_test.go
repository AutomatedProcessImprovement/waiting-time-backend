package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

func makeTestApplication() (*Application, error) {
	config := DefaultConfiguration()
	config.AssetsDir = "../assets"
	config.ResultsDir = "../assets/results"
	return NewApplication(config)
}

func TestGetMethods(t *testing.T) {
	// setup

	app, err := makeTestApplication()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	// test cases

	tests := []struct {
		name          string
		method        string
		path          string
		input         interface{}
		output        interface{}
		outputDecoded interface{}
		statusCode    int
		contentType   string
	}{
		{
			name:          "static asset",
			method:        "GET",
			path:          "/assets/samples/manual_log_5.csv",
			input:         nil,
			output:        nil,
			outputDecoded: nil,
			statusCode:    http.StatusOK,
			contentType:   "text/csv; charset=utf-8",
		},
		{
			name:          "static asset not found",
			method:        "GET",
			path:          "/foobar.csv",
			input:         nil,
			output:        nil,
			outputDecoded: nil,
			statusCode:    http.StatusNotFound,
			contentType:   "text/plain; charset=utf-8",
		},
		{
			name:          "swagger.json",
			method:        "GET",
			path:          "/swagger.json",
			input:         nil,
			output:        nil,
			outputDecoded: nil,
			statusCode:    http.StatusOK,
			contentType:   "application/json; charset=utf-8",
		},
		{
			name:          "get jobs",
			method:        "GET",
			path:          "/jobs",
			input:         nil,
			output:        nil,
			outputDecoded: []*model.ApiJobsResponse{},
			statusCode:    http.StatusOK,
			contentType:   "application/json; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(app.GetRouter())

			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != tt.statusCode {
				t.Fatalf("expected status code %d, got %d", tt.statusCode, res.StatusCode)
			}

			if strings.ToLower(res.Header.Get("Content-Type")) != strings.ToLower(tt.contentType) {
				t.Fatalf("expected content type %s, got %s", tt.contentType, res.Header.Get("Content-Type"))
			}

			if tt.outputDecoded != nil {
				err := json.NewDecoder(res.Body).Decode(&tt.outputDecoded)
				if err != nil {
					t.Fatal(err)
				}
			}

			ts.Close()
		})
	}
}

func TestPostJob(t *testing.T) {
	// setup

	app, err := makeTestApplication()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	b, err := ioutil.ReadFile("../assets/samples/manual_log_5.csv")
	if err != nil {
		t.Fatal(err)
	}

	// test cases

	tests := []struct {
		name                string
		method              string
		path                string
		input               []byte
		output              interface{}
		outputDecoded       *model.ApiSingleJobResponse
		statusCode          int
		inputContentType    string
		expectedContentType string
		expectedError       string
	}{
		{
			name:                "post job",
			method:              "POST",
			path:                "/jobs",
			input:               b,
			output:              nil,
			outputDecoded:       &model.ApiSingleJobResponse{},
			statusCode:          http.StatusCreated,
			inputContentType:    "text/csv; charset=utf-8",
			expectedContentType: "application/json; charset=utf-8",
		},
		{
			name:                "post job no bytes",
			method:              "POST",
			path:                "/jobs",
			input:               []byte{},
			output:              nil,
			outputDecoded:       &model.ApiSingleJobResponse{},
			statusCode:          http.StatusBadRequest,
			inputContentType:    "text/csv; charset=utf-8",
			expectedContentType: "application/json; charset=utf-8",
			expectedError:       "failed to create a job from the request body; request body is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(app.GetRouter())

			req, err := http.NewRequest(tt.method, ts.URL+tt.path, bytes.NewReader(tt.input))
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", tt.inputContentType)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != tt.statusCode {
				t.Fatalf("expected status code %d, got %d", tt.statusCode, res.StatusCode)
			}

			if strings.ToLower(res.Header.Get("Content-Type")) != strings.ToLower(tt.expectedContentType) {
				t.Fatalf("expected content type %s, got %s", tt.expectedContentType, res.Header.Get("Content-Type"))
			}

			if tt.outputDecoded != nil {
				err := json.NewDecoder(res.Body).Decode(&tt.outputDecoded)
				if err != nil {
					t.Fatal(err)
				}

				if tt.outputDecoded.Error != tt.expectedError {
					t.Fatalf("expected error %s, got %s", tt.expectedError, tt.outputDecoded.Error)
				}

				jobDir := path.Join(app.config.ResultsDir, tt.outputDecoded.Job.ID)

				if err = os.RemoveAll(jobDir); err != nil {
					t.Fatal(err)
				}
			}

			ts.Close()
		})
	}
}
