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
	Loaded               bool             `json:"loaded"`
	SimulationTime       float64          `json:"time"`
	RealTimeFactor       float64          `json:"realTimeFactor"`
	IsRealTimeSimulation bool             `json:"isRealTime"`
	ConfigDir            string           `json:"configDir,omitempty"`
	Status               string           `json:"status,omitempty"`
	Module               Module           `json:"module,omitempty"`
	Trends               []Trend          `json:"trends"`
	ModuleData           *MetaData        `json:"module-data,omitempty"`
	Feedback             *CommandFeedback `json:"feedback,omitempy"`
}

type TrendSignal struct {
	Module          string    `json:"module"`
	SlaveIndex      int       `json:"slave-index"`
	Signal          string    `json:"signal"`
	Causality       string    `json:"causality"`
	Type            string    `json:"type"`
	ValueReference  int       `json:"value-reference"`
	TrendValues     []float64 `json:"values,omitempty"`
	TrendTimestamps []float64 `json:"timestamps,omitempty"`
}

type Trend struct {
	Id           int           `json:"id"`
	PlotType     string        `json:"plot-type"`
	Label        string        `json:"label"`
	TrendSignals []TrendSignal `json:"trend-values"`
}

type TrendSpec struct {
	Begin float64
	End   float64
	Range float64
	Auto  bool
}

type SimulationStatus struct {
	Loaded              bool
	ConfigDir           string
	Module              string
	SignalSubscriptions []Variable
	Trends              []Trend
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

type CommandFeedback struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command"`
}
