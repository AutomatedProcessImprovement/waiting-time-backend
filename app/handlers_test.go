package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticAssets(t *testing.T) {
	// setup

	config := DefaultConfiguration()
	config.AssetsDir = "../assets"
	config.ResultsDir = "../assets/results"
	app, err := NewApplication(config)
	if err != nil {
		t.Fatal(err)
	}

	// test cases

	tests := []struct {
		name        string
		method      string
		path        string
		input       interface{}
		output      interface{}
		statusCode  int
		contentType string
	}{
		{
			name:        "static asset",
			method:      "GET",
			path:        "/assets/samples/manual_log_5.csv",
			input:       nil,
			output:      nil,
			statusCode:  http.StatusOK,
			contentType: "text/csv; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			StaticAssets(app).ServeHTTP(w, request)

			response := w.Result()

			if response.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, w.Code)
			}

			if response.Header.Get("Content-Type") != tt.contentType {
				t.Errorf("expected content type %s, got %s", tt.contentType, response.Header.Get("Content-Type"))
			}
		})
	}

	// teardown

	app.Close()
}
