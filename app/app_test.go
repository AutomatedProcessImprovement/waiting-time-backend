package app

import (
	"testing"
)

func Test_jobResultFromPath(t *testing.T) {
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
		name       string
		reportPath string
	}{
		{
			name:       "valid result",
			reportPath: "../assets/tests/manual_log_5_transitions_report.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := app.jobResultFromPath(tt.reportPath)
			if err != nil {
				t.Fatal(err)
			}

			if result.CTEImpact == nil {
				t.Fatal("CTEImpact is nil")
			}

			if len(result.Report) == 0 {
				t.Fatal("Report is empty")
			}

			if len(result.Report[0].WtByResource) == 0 {
				t.Fatal("WtByResource is empty")
			}

			if result.Report[0].CTEImpact == nil {
				t.Fatal("CTEImpactPerWt is nil")
			}

			if result.Report[0].WtByResource[0].CTEImpact == nil {
				t.Fatal("CTEImpactPerWt is nil")
			}
		})
	}
}
