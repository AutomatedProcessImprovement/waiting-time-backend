package model

type JobResultReportItem struct {
	SourceActivity   string                  `json:"source_activity"`
	TargetActivity   string                  `json:"target_activity"`
	CaseFreq         float64                 `json:"case_freq"`
	TotalFreq        float64                 `json:"total_freq"`
	TotalWt          float64                 `json:"total_wt"`
	BatchingWt       float64                 `json:"batching_wt"`
	PrioritizationWt float64                 `json:"prioritization_wt"`
	ContentionWt     float64                 `json:"contention_wt"`
	UnavailabilityWt float64                 `json:"unavailability_wt"`
	ExtraneousWt     float64                 `json:"extraneous_wt"`
	WtByResource     []JobResultResourceItem `json:"wt_by_resource"`
}
