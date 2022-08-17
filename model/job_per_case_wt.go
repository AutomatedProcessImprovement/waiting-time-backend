package model

// JobPerCaseWT displays measures per case.
//
// swagger:model
type JobPerCaseWT struct {
	CaseID    float64 `json:"case:concept:name"`
	CasePT    float64 `json:"pt_total"`
	CaseWT    float64 `json:"wt_total"`
	CTEImpact float64 `json:"cte_impact"`
}
