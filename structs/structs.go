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
	Module               Module        `json:"module,omitempty"`
	TrendSignals         []TrendSignal `json:"trend-values"`
	ModuleData           *MetaData     `json:"module-data,omitempty"`
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
	Mutex               sync.Mutex
	Loaded              bool
	ConfigDir           string
	Module              string
	SignalSubscriptions []Variable
	TrendSignals        []TrendSignal
	TrendSpec           TrendSpec
	Status              string
	MetaChan            chan *MetaData
}

type Variable struct {
	Name           string `json:"name"`
	ValueReference int    `json:"value-reference"`
	Causality      string `json:"causality"`
	Variability    string `json:"variability"`
	Type           string `json:"type"`
}

type FMU struct {
	Name           string     `json:"name"`
	ExecutionIndex int        `json:"index"`
	Variables      []Variable `json:"variables"`
}

type MetaData struct {
	FMUs []FMU `json:"fmus"`
}
