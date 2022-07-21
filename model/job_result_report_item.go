package model

type JobResultReportItem struct {
	SourceActivity   string                  `json:"source_activity,omitempty"`
	TargetActivity   string                  `json:"target_activity,omitempty"`
	CaseFreq         float64                 `json:"case_freq,omitempty"`
	TotalFreq        float64                 `json:"total_freq,omitempty"`
	TotalWt          float64                 `json:"total_wt,omitempty"`
	BatchingWt       float64                 `json:"batching_wt,omitempty"`
	PrioritizationWt float64                 `json:"prioritization_wt,omitempty"`
	ContentionWt     float64                 `json:"contention_wt,omitempty"`
	UnavailabilityWt float64                 `json:"unavailability_wt,omitempty"`
	ExtraneousWt     float64                 `json:"extraneous_wt,omitempty"`
	WtByResource     []JobResultResourceItem `json:"wt_by_resource,omitempty"`
}
