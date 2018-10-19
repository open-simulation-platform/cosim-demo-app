package structs

type Signal struct {
	Name      string      `json:"name"`
	Causality string      `json:"causality"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
}

type Module struct {
	Signals []Signal `json:"signals,omitempty"`
	Name    string   `json:"name,omitempty"`
}

type JsonResponse struct {
	Loaded       bool          `json:"loaded"`
	Status       string        `json:"status,omitempty"`
	Modules      []string      `json:"modules,omitempty"`
	Module       Module        `json:"module,omitempty"`
	TrendSignals []TrendSignal `json:"trendSignals,omitempty"`
}

type TrendSignal struct {
	Module          string
	Signal          string
	TrendValues     []float64
	TrendTimestamps []int
}

type SimulationStatus struct {
	Loaded       bool
	Module       Module
	TrendSignals []TrendSignal
	Status       string
}

type Variable struct {
	Name           string
	ValueReference int
	Causality      string
	Variability    string
	Type           string
}

type FMU struct {
	Name           string
	ExecutionIndex int
	ObserverIndex  int
	Variables      []Variable
}

type MetaData struct {
	FMUs []FMU
}
