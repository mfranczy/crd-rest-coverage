package stats

// Coverage represents a REST API statistics
type Coverage struct {
	UniqueHits         int                             `json:"uniqueHits"`
	ExpectedUniqueHits int                             `json:"expectedUniqueHits"`
	Percent            float64                         `json:"percent"`
	Endpoints          map[string]map[string]*Endpoint `json:"endpoints"`
}

// Endpoint represents a basic statistics structure which is used to calculate REST API coverage
type Endpoint struct {
	ParamsHitsDetails  `json:"paramsHitsDetails"`
	UniqueHits         int     `json:"uniqueHits"`
	ExpectedUniqueHits int     `json:"expectedUniqueHits"`
	Percent            float64 `json:"percent"`
	MethodCalled       bool    `json:"methodCalled"`
	Path               string  `json:"path"`
	Method             string  `json:"method"`
}

// ParamsHitsDetails represents a parameter path with its occurence number
type ParamsHitsDetails struct {
	Body  map[string]int `json:"body"`
	Query map[string]int `json:"query"`
}
