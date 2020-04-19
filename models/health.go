package models

import "time"

type Health struct {
	CurrentTime       time.Time          `json:"currentTime"`
	StartTime         time.Time          `json:"startTime"`
	NetworkInterfaces []string           `json:"networkInterfaces"`
	Connections       []ConnectionStatus `json:"connections"`
}

type ConnectionStatus struct {
	ServiceName string   `json:"serviceName"`
	ActiveNodes []string `json:"activeNodes"`
	DownNodes   []string `json:"downNodes,omitempty"`
}
