package model

type JobResultReportItem struct {
	SourceActivity   string                  `json:"source_activity,omitempty"`
	TargetActivity   string                  `json:"target_activity,omitempty"`
	CaseFreq         uint64                  `json:"case_freq,omitempty"`
	TotalFreq        uint64                  `json:"total_freq,omitempty"`
	TotalWt          uint64                  `json:"total_wt,omitempty"`
	BatchingWt       uint64                  `json:"batching_wt,omitempty"`
	PrioritizationWt uint64                  `json:"prioritization_wt,omitempty"`
	ContentionWt     uint64                  `json:"contention_wt,omitempty"`
	UnavailabilityWt uint64                  `json:"unavailability_wt,omitempty"`
	ExtraneousWt     uint64                  `json:"extraneous_wt,omitempty"`
	WtByResource     []JobResultResourceItem `json:"wt_by_resource,omitempty"`
}
