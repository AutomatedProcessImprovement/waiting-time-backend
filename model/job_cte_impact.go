package model

// JobCteImpact represents CTE impact of waiting times on the process level.
//
// swagger:model
type JobCteImpact struct {
	BatchingImpact       float64 `json:"batching_impact,omitempty"`
	ContentionImpact     float64 `json:"contention_impact,omitempty"`
	PrioritizationImpact float64 `json:"prioritization_impact,omitempty"`
	UnavailabilityImpact float64 `json:"unavailability_impact,omitempty"`
	ExtraneousImpact     float64 `json:"extraneous_impact,omitempty"`
}
