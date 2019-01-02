package structs

import (
	"sync"
)

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
	Loaded               bool          `json:"loaded"`
	SimulationTime       float64       `json:"time"`
	RealTimeFactor       float64       `json:"realTimeFactor"`
	IsRealTimeSimulation bool          `json:"isRealTime"`
	ConfigDir            string        `json:"configDir,omitempty"`
	Status               string        `json:"status,omitempty"`
	Modules              []string      `json:"modules"`
	Module               Module        `json:"module,omitempty"`
	TrendSignals         []TrendSignal `json:"trend-values"`
}

type TrendSignal struct {
	Module          string    `json:"module"`
	Signal          string    `json:"signal"`
	Causality       string    `json:"causality"`
	Type            string    `json:"type"`
	TrendValues     []float64 `json:"values,omitempty"`
	TrendTimestamps []float64 `json:"labels,omitempty"`
}

type TrendSpec struct {
	Begin float64
	End   float64
	Range float64
	Auto  bool
}

type SimulationStatus struct {
	Mutex        sync.Mutex
	Loaded       bool
	ConfigDir    string
	Module       Module
	TrendSignals []TrendSignal
	TrendSpec    TrendSpec
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
	Variables      []Variable
}

type MetaData struct {
	FMUs []FMU
}
