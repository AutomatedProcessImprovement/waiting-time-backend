package model

type JobResult struct {
	NumCases               float64                `json:"num_cases,omitempty"`
	NumActivities          float64                `json:"num_activities,omitempty"`
	NumActivityInstances   float64                `json:"num_activity_instances,omitempty"`
	NumTransitions         float64                `json:"num_transitions,omitempty"`
	NumTransitionInstances float64                `json:"num_transition_instances,omitempty"`
	TotalWt                float64                `json:"total_wt,omitempty"`
	TotalBatchingWt        float64                `json:"total_batching_wt,omitempty"`
	TotalPrioritizationWt  float64                `json:"total_prioritization_wt,omitempty"`
	TotalContentionWt      float64                `json:"total_contention_wt,omitempty"`
	TotalUnavailabilityWt  float64                `json:"total_unavailability_wt,omitempty"`
	TotalExtraneousWt      float64                `json:"total_extraneous_wt,omitempty"`
	Report                 []*JobResultReportItem `json:"report,omitempty"`
	CTEImpact              *JobCteImpact          `json:"cte_impact,omitempty"`
}
