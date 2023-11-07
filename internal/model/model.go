package model

import (
	"regexp"
	"time"
)

type LogLevelType string

const (
	Trace   string = "trace"
	Info    string = "info"
	Debug   string = "debug"
	Warning string = "warning"
	Error   string = "error"
	Fatal   string = "fatal"
)

type Cluster struct {
	Config     string   `json:"config"`
	Name       string   `json:"name"`
	Namespaces []string `json:"namespaces"`
}

type ServiceScan struct {
	ServiceName        string         `json:"service_name"`
	Uptime             time.Duration  `json:"uptime"`
	RestartsCount      int            `json:"restarts_count"`
	LogTypeCountMap    map[string]int `json:"logs_info"`
	NoneJsonLinesCount int            `json:"none_json_lines_count"`
	TotalLines         int            `json:"total_lines"`
	ScanFinishTime     time.Time      `json:"scan_finish_time"`
}

type JobScan struct {
	JobName        string        `json:"job_name"`
	Age            time.Duration `json:"age"`
	FullLog        string        `json:"full_log"`
	GrepPattern    regexp.Regexp `json:"grep_pattern"`
	GrepLog        []string      `json:"grep_log"`
	ScanFinishTime time.Time     `json:"scan_finish_time"`
}

type CommonServiceLog struct {
	Level LogLevelType `json:"level"`
}
