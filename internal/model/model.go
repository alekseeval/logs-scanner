package model

import (
	"regexp"
	"time"
)

type LogLevelType string

const (
	Trace   LogLevelType = "trace"
	Info    LogLevelType = "info"
	Debug   LogLevelType = "debug"
	Warning LogLevelType = "warning"
	Error   LogLevelType = "error"
	Fatal   LogLevelType = "fatal"
)

type KubeConfig struct {
	Config     string
	Name       string
	NameSpaces []string
}

type ServiceScan struct {
	ServiceName        string
	Uptime             time.Duration
	RestartsCount      int
	LogTypeCountMap    map[LogLevelType]int
	NoneJsonLinesCount int
	TotalLines         int
	ScanFinishTime     time.Time
}

type JobScan struct {
	JobName        string
	Age            time.Duration
	FullLog        string
	GrepPattern    regexp.Regexp
	GrepLog        []string
	ScanFinishTime time.Time
}

type CommonServiceLog struct {
	Level LogLevelType `json:"level"`
}
