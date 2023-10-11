package model

import "time"

type JobResultItem struct {
	CaseID              string    `json:"case_id" db:"case_id"`
	WtTotal             float64   `json:"wt_total" db:"wt_total"`
	WtExtraneous        float64   `json:"wt_extraneous" db:"wt_extraneous"`
	StartTime           time.Time `json:"start_time" db:"start_time"`
	SourceActivity      string    `json:"source_activity" db:"source_activity"`
	SourceResource      string    `json:"source_resource" db:"source_resource"`
	DestinationActivity string    `json:"destination_activity" db:"destination_activity"`
	WtPrioritization    float64   `json:"wt_prioritization" db:"wt_prioritization"`
	WtUnavailability    float64   `json:"wt_unavailability" db:"wt_unavailability"`
	EndTime             time.Time `json:"end_time" db:"end_time"`
	DestinationResource string    `json:"destination_resource" db:"destination_resource"`
	WtContention        float64   `json:"wt_contention" db:"wt_contention"`
	WtBatching          float64   `json:"wt_batching" db:"wt_batching"`
}
