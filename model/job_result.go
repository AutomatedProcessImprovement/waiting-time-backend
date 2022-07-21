package model

// JobResult is a result of a job's execution which contains a summary of the transitions analysis report, a report
// itself and CTE impact of waiting times on the process level and on a transition level.
//
// swagger:model
type JobResult struct {
	NumCases               float64                `json:"num_cases"`
	NumActivities          float64                `json:"num_activities"`
	NumActivityInstances   float64                `json:"num_activity_instances"`
	NumTransitions         float64                `json:"num_transitions"`
	NumTransitionInstances float64                `json:"num_transition_instances"`
	TotalWt                float64                `json:"total_wt"`
	TotalBatchingWt        float64                `json:"total_batching_wt"`
	TotalPrioritizationWt  float64                `json:"total_prioritization_wt"`
	TotalContentionWt      float64                `json:"total_contention_wt"`
	TotalUnavailabilityWt  float64                `json:"total_unavailability_wt"`
	TotalExtraneousWt      float64                `json:"total_extraneous_wt"`
	Report                 []*JobResultReportItem `json:"report"`
	CTEImpact              *JobCteImpact          `json:"cte_impact"`
}
