package model

type JobResult struct {
	NumCases               uint64                `json:"num_cases,omitempty"`
	NumActivities          uint64                `json:"num_activities,omitempty"`
	NumActivityInstances   uint64                `json:"num_activity_instances,omitempty"`
	NumTransitions         uint64                `json:"num_transitions,omitempty"`
	NumTransitionInstances uint64                `json:"num_transition_instances,omitempty"`
	TotalWt                uint64                `json:"total_wt,omitempty"`
	TotalBatchingWt        uint64                `json:"total_batching_wt,omitempty"`
	TotalPrioritizationWt  uint64                `json:"total_prioritization_wt,omitempty"`
	TotalContentionWt      uint64                `json:"total_contention_wt,omitempty"`
	TotalUnavailabilityWt  uint64                `json:"total_unavailability_wt,omitempty"`
	TotalExtraneousWt      uint64                `json:"total_extraneous_wt,omitempty"`
	Report                 []JobResultReportItem `json:"report,omitempty"`
}
