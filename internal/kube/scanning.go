package kube

import (
	"bufio"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"scan_project/internal/model"
	"strings"
	"time"
)

func (ks *KubeScanner) scanServiceLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (serviceScan *model.ServiceScan, err error) {
	serviceScan = &model.ServiceScan{
		ServiceName:     pod.Name,
		LogTypeCountMap: make(map[model.LogLevelType]int),
	}
	serviceScan.ServiceName = pod.Name
	// Use first pod container, to get restarts count
	var restartCount int
	if len(pod.Status.ContainerStatuses) != 0 {
		restartCount = int(pod.Status.ContainerStatuses[0].RestartCount)
	}
	serviceScan.RestartsCount = restartCount
	serviceScan.Uptime = time.Now().Sub(pod.CreationTimestamp.Time)
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close()
	scanner := bufio.NewScanner(podLogsStream)
	linesCount := 0
	for scanner.Scan() {
		linesCount++
		log := &model.CommonServiceLog{}
		err := json.Unmarshal(scanner.Bytes(), log)
		if err != nil {
			serviceScan.NoneJsonLinesCount++
			continue
		}
		switch log.Level {
		case model.Trace, model.Debug, model.Info, model.Warning, model.Error, model.Fatal:
			serviceScan.LogTypeCountMap[log.Level] += 1
		default:
			ks.logger.Warning(fmt.Sprintf("Unknown log level -- %s", log.Level))
		}
	}
	serviceScan.TotalLines = linesCount
	serviceScan.ScanFinishTime = time.Now()
	return serviceScan, nil
}

func (ks *KubeScanner) scanJobLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (*model.JobScan, error) {
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.TODO())
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
