// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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

type ManipulatedVariable struct {
	SlaveIndex     int    `json:"slaveIndex"`
	Type           string `json:"type"`
	ValueReference int    `json:"valueReference"`
}

type JsonResponse struct {
	Loading              bool                  `json:"loading"`
	Loaded               bool                  `json:"loaded"`
	ExecutionState       string                `json:"executionState"`
	LastErrorCode        string                `json:"lastErrorCode"`
	LastErrorMessage     string                `json:"lastErrorMessage"`
	SimulationTime       float64               `json:"time"`
	RealTimeFactor       float64               `json:"realTimeFactor"`
	RealTimeFactorTarget float64               `json:"realTimeFactorTarget"`
	IsRealTimeSimulation bool                  `json:"isRealTime"`
	ConfigDir            string                `json:"configDir,omitempty"`
	Status               string                `json:"status,omitempty"`
	Module               Module                `json:"module,omitempty"`
	Trends               []Trend               `json:"trends"`
	ModuleData           *MetaData             `json:"module-data,omitempty"`
	Feedback             *CommandFeedback      `json:"feedback,omitempy"`
	Scenarios            *[]string             `json:"scenarios,omitempty"`
	Scenario             *interface{}          `json:"scenario,omitempty"`
	RunningScenario      string                `json:"running-scenario"`
	ManipulatedVariables []ManipulatedVariable `json:"manipulatedVariables"`
}

type TrendSignal struct {
	Module         string    `json:"module"`
	SlaveIndex     int       `json:"slave-index"`
	Signal         string    `json:"signal"`
	Causality      string    `json:"causality"`
	Type           string    `json:"type"`
	ValueReference int       `json:"value-reference"`
	TrendXValues   []float64 `json:"xvals,omitempty"`
	TrendYValues   []float64 `json:"yvals,omitempty"`
}

type Trend struct {
	Id           int           `json:"id"`
	PlotType     string        `json:"plot-type"`
	Label        string        `json:"label"`
	TrendSignals []TrendSignal `json:"trend-values"`
	Spec         TrendSpec     `json:"spec"`
}

type TrendSpec struct {
	Begin float64 `json:"begin"`
	End   float64 `json:"end"`
	Range float64 `json:"range"`
	Auto  bool    `json:"auto"`
}

type ShortLivedData struct {
	Scenarios  *[]string
	Scenario   *interface{}
	ModuleData *MetaData
}

type SimulationStatus struct {
	Loading             bool
	Loaded              bool
	ConfigDir           string
	Module              string
	SignalSubscriptions []Variable
	Trends              []Trend
	Status              string
	CurrentScenario     string
	ActiveTrend         int
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

type CommandFeedback struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command"`
}

type PlotVariable struct {
	Simulator string `json:"simulator"`
	Variable  string `json:"variable"`
}

type Plot struct {
	Label         string         `json:"label"`
	PlotType      string         `json:"plotType"`
	PlotVariables []PlotVariable `json:"variables"`
}

type PlotConfig struct {
	Plots []Plot `json:"plots"`
}
