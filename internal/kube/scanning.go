package kube

import (
	"bufio"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"scan_project/internal/model"
	"strings"
	"time"
)

func (ks *KubeScanner) scanServiceLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (serviceScan *model.ServiceScan, err error) {
	// Init pod common data into Scan struct
	serviceScan = &model.ServiceScan{
		ServiceName:     pod.Name,
		LogTypeCountMap: make(map[string]int),
		Uptime:          time.Now().Sub(pod.CreationTimestamp.Time),
	}
	var restartCount int
	if len(pod.Status.ContainerStatuses) != 0 {
		restartCount = int(pod.Status.ContainerStatuses[0].RestartCount) // Use first pod container, to get restarts count
	}
	serviceScan.RestartsCount = restartCount
	// Get all pod logs
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close()
	scanner := bufio.NewScanner(podLogsStream)
	linesCount := 0
	for scanner.Scan() {
		linesCount++
		logBytes := scanner.Bytes()
		foundLevelBytes := ks.servicesRegexp.FindSubmatch(logBytes)
		if foundLevelBytes == nil {
			serviceScan.NoneJsonLinesCount++
			continue
		}
		if len(foundLevelBytes) != 1 {
			ks.logger.
				WithField("log", string(logBytes)).
				Warning("Several \"level\" key founds in log")
			continue
		}
		foundLevelStr := string(foundLevelBytes[0])
		level := foundLevelStr[9 : len(foundLevelStr)-1]
		switch level {
		case model.Trace, model.Debug, model.Info, model.Warning, model.Error, model.Fatal:
			serviceScan.LogTypeCountMap[level] += 1
		default:
			ks.logger.Warning(fmt.Sprintf("Unknown log level -- %s", level))
		}
	}
	serviceScan.TotalLines = linesCount
	serviceScan.ScanFinishTime = time.Now()
	return serviceScan, nil
}

func (ks *KubeScanner) scanJobLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (*model.JobScan, error) {
	// Get all pod logs
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close()
	var sb strings.Builder
	matchedLogRows := make([]string, 0)
	scanner := bufio.NewScanner(podLogsStream)
	for scanner.Scan() {
		strokeText := scanner.Text()
		if ks.jobsRegexp.MatchString(strokeText) {
			matchedLogRows = append(matchedLogRows, strokeText)
		}
		sb.WriteString(scanner.Text())
		sb.WriteRune('\n')
	}
	return &model.JobScan{
		JobName:        pod.Name,
		Age:            time.Now().Sub(pod.CreationTimestamp.Time),
		FullLog:        sb.String(),
		GrepPattern:    *ks.jobsRegexp,
		GrepLog:        matchedLogRows,
		ScanFinishTime: time.Now(),
	}, nil
}
