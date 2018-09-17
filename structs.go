package main

type JsonRequest struct {
	Command     string `json:"command,omitempty"`
	Module      string `json:"module,omitempty"`
	Modules     bool   `json:"modules,omitempty"`
	Connections bool   `json:"connections,omitempty"`
}

type Signal struct {
	Name  string  `json:"name,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type Module struct {
	Signals []Signal `json:"signals,omitempty"`
	Name    string   `json:"name,omitempty"`
}

type JsonResponse struct {
	Modules []string `json:"modules,omitempty"`
	Module  Module   `json:"module,omitempty"`
}
