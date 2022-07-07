package app

type JobResultResourceItem struct {
	SourceResource   string `json:"source_resource,omitempty"`
	TargetResource   string `json:"target_resource,omitempty"`
	CaseFreq         uint64 `json:"case_freq,omitempty"`
	TotalFreq        uint64 `json:"total_freq,omitempty"`
	TotalWt          uint64 `json:"total_wt,omitempty"`
	BatchingWt       uint64 `json:"batching_wt,omitempty"`
	PrioritizationWt uint64 `json:"prioritization_wt,omitempty"`
	ContentionWt     uint64 `json:"contention_wt,omitempty"`
	UnavailabilityWt uint64 `json:"unavailability_wt,omitempty"`
	ExtraneousWt     uint64 `json:"extraneous_wt,omitempty"`
}
