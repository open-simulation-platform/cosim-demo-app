package main

type JsonRequest struct {
	Command     string `json:"command,omitempty"`
	Module      string `json:"module,omitempty"`
	Modules     bool   `json:"modules,omitempty"`
	Connections bool   `json:"connections,omitempty"`
}

type JsonResponse struct {
	Modules     []string `json:"modules,omitempty"`
	Status      string   `json:"status,omitempty"`
	SignalValue float64  `json:"signalValue,omitempty"`
}

