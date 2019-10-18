package dex

import (
	"time"
)

// Metric struct contains all the fields required
// to store the statistics for a given request
type Metric struct {
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Platform string    `json:"platform"`
	Path     string    `json:"path"`
	Host     string    `json:"host"`
	Type     string    `json:"type"`
	Value    int64     `json:"value"`
}

// NewMetric is used to create a new metric struct
// using the passed values
func NewMetric(now time.Time, hostname, path, host, mtype string, value int64) Metric {
	return Metric{
		Time:     now,
		Hostname: hostname,
		Platform: "go",
		Path:     path,
		Host:     host,
		Type:     mtype,
		Value:    value,
	}
}
