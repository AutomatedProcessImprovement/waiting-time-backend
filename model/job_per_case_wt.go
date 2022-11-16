package model

// JobPerCaseWT displays measures per case.
//
// swagger:model
type JobPerCaseWT struct {
	CaseID    string  `json:"case_id"`
	CasePT    float64 `json:"pt_total"`
	CaseWT    float64 `json:"wt_total"`
	CTEImpact float64 `json:"cte_impact"`
}
