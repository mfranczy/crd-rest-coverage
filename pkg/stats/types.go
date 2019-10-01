package stats

// Coverage represents a REST API statistics
type Coverage struct {
	Total     float64                         `json:"total"`
	Endpoints map[string]map[string]*Endpoint `json:"endpoints"`
}

// Endpoint represents a basic statistics structure which is used to calculate REST API coverage
type Endpoint struct {
	ParamsHitDetails `json:"paramsHitDetails"`
	UniqueHit        int     `json:"hit"`
	Sum              int     `json:"sum"`
	MethodCalled     bool    `json:"methodCalled"`
	Path             string  `json:"path"`
	Method           string  `json:"method"`
	Total            float64 `json:"total"`
}

// ParamsHitDetails represents a parameter path with its occurence number
type ParamsHitDetails struct {
	Body  map[string]int `json:"body"`
	Query map[string]int `json:"query"`
}
