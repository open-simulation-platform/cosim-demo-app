package structs

import (
	"github.com/shirou/gopsutil/mem"
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
	Loaded         bool     `json:"loaded"`
	SimulationTime float64  `json:"time"`
	ConfigDir      string   `json:"configDir,omitempty"`
	Status         string   `json:"status,omitempty"`
	Modules        []string `json:"modules"`
	Module         Module   `json:"module,omitempty"`
	Memory         *mem.VirtualMemoryStat
	TrendSignals   []TrendSignal `json:"trendSignals,omitempty"`
}

type TrendSignal struct {
	Module          string
	Signal          string
	Causality       string
	Type            string
	TrendValues     []float64
	TrendTimestamps []float64
}

type TrendSpec struct {
	Begin float64
	End   float64
	Range float64
	Auto  bool
}

type SimulationStatus struct {
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
