package model

type JobResultResourceItem struct {
	SourceResource   string  `json:"source_resource,omitempty"`
	TargetResource   string  `json:"target_resource,omitempty"`
	CaseFreq         float64 `json:"case_freq,omitempty"`
	TotalFreq        float64 `json:"total_freq,omitempty"`
	TotalWt          float64 `json:"total_wt,omitempty"`
	BatchingWt       float64 `json:"batching_wt,omitempty"`
	PrioritizationWt float64 `json:"prioritization_wt,omitempty"`
	ContentionWt     float64 `json:"contention_wt,omitempty"`
	UnavailabilityWt float64 `json:"unavailability_wt,omitempty"`
	ExtraneousWt     float64 `json:"extraneous_wt,omitempty"`
}
